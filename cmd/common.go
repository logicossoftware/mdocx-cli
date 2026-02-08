package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/logicossoftware/go-mdocx"
)

type headerInfo struct {
	MagicHex       string `json:"magic_hex"`
	MagicValid     bool   `json:"magic_valid"`
	Version        uint16 `json:"version"`
	HeaderFlags    uint16 `json:"header_flags"`
	FixedHdrSize   uint32 `json:"fixed_header_size"`
	MetadataLength uint32 `json:"metadata_length"`
	ReservedClean  bool   `json:"reserved_clean"`
}

func readHeaderInfo(path string) (*headerInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var buf [32]byte
	if _, err := io.ReadFull(f, buf[:]); err != nil {
		return nil, err
	}

	expectedMagic := [8]byte{'M', 'D', 'O', 'C', 'X', '\r', '\n', 0x1A}
	var actualMagic [8]byte
	copy(actualMagic[:], buf[0:8])
	magicValid := actualMagic == expectedMagic

	// Check that reserved bytes 20-31 are all zero.
	reservedClean := true
	for _, b := range buf[20:32] {
		if b != 0 {
			reservedClean = false
			break
		}
	}

	return &headerInfo{
		MagicHex:       hex.EncodeToString(buf[0:8]),
		MagicValid:     magicValid,
		Version:        uint16(buf[8]) | uint16(buf[9])<<8,
		HeaderFlags:    uint16(buf[10]) | uint16(buf[11])<<8,
		FixedHdrSize:   uint32(buf[12]) | uint32(buf[13])<<8 | uint32(buf[14])<<16 | uint32(buf[15])<<24,
		MetadataLength: uint32(buf[16]) | uint32(buf[17])<<8 | uint32(buf[18])<<16 | uint32(buf[19])<<24,
		ReservedClean:  reservedClean,
	}, nil
}

func parseCompression(value string) (mdocx.Compression, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "zstd":
		return mdocx.CompZSTD, nil
	case "none":
		return mdocx.CompNone, nil
	case "zip":
		return mdocx.CompZIP, nil
	case "lz4":
		return mdocx.CompLZ4, nil
	case "br", "brotli":
		return mdocx.CompBR, nil
	default:
		return mdocx.CompNone, fmt.Errorf("unknown compression: %s", value)
	}
}

func readMetadataJSON(path string) (map[string]any, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	if out == nil {
		return nil, errors.New("metadata must be a JSON object")
	}
	return out, nil
}

func collectMarkdownFiles(inputs []string, baseDir string) ([]mdocx.MarkdownFile, error) {
	var files []string
	if baseDir != "" {
		collected, err := collectFiles(baseDir, func(rel string, info os.DirEntry) bool {
			return !info.IsDir() && strings.HasSuffix(strings.ToLower(rel), ".md")
		})
		if err != nil {
			return nil, err
		}
		for _, rel := range collected {
			files = append(files, filepath.Join(baseDir, rel))
		}
	}
	for _, in := range inputs {
		stat, err := os.Stat(in)
		if err != nil {
			return nil, err
		}
		if stat.IsDir() {
			collected, err := collectFiles(in, func(rel string, info os.DirEntry) bool {
				return !info.IsDir() && strings.HasSuffix(strings.ToLower(rel), ".md")
			})
			if err != nil {
				return nil, err
			}
			for _, rel := range collected {
				files = append(files, filepath.Join(in, rel))
			}
			continue
		}
		files = append(files, in)
	}

	if len(files) == 0 {
		return nil, errors.New("no markdown files found")
	}

	cwd, _ := os.Getwd()
	seen := make(map[string]struct{})
	out := make([]mdocx.MarkdownFile, 0, len(files))
	for _, filePath := range files {
		b, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		containerPath := containerPathFromFile(cwd, filePath)
		if _, ok := seen[containerPath]; ok {
			return nil, fmt.Errorf("duplicate markdown path: %s", containerPath)
		}
		seen[containerPath] = struct{}{}
		out = append(out, mdocx.MarkdownFile{Path: containerPath, Content: b})
	}

	return out, nil
}

func collectMediaItems(mediaDir string) ([]mdocx.MediaItem, error) {
	if strings.TrimSpace(mediaDir) == "" {
		return nil, nil
	}
	collected, err := collectFiles(mediaDir, func(rel string, info os.DirEntry) bool {
		return !info.IsDir()
	})
	if err != nil {
		return nil, err
	}
	seenIDs := make(map[string]string) // id -> original path
	out := make([]mdocx.MediaItem, 0, len(collected))
	for _, rel := range collected {
		fsPath := filepath.Join(mediaDir, rel)
		b, err := os.ReadFile(fsPath)
		if err != nil {
			return nil, err
		}
		containerPath := filepath.ToSlash(rel)
		id := makeIDFromPath(containerPath)
		if prevPath, ok := seenIDs[id]; ok {
			return nil, fmt.Errorf("duplicate media ID %q generated from %q (conflicts with %q)", id, containerPath, prevPath)
		}
		seenIDs[id] = containerPath
		m := detectMimeType(fsPath)
		out = append(out, mdocx.MediaItem{
			ID:       id,
			Path:     containerPath,
			MIMEType: m,
			Data:     b,
			SHA256:   sha256.Sum256(b),
		})
	}
	return out, nil
}

func collectFiles(root string, keep func(rel string, info os.DirEntry) bool) ([]string, error) {
	var out []string
	if err := filepath.WalkDir(root, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if keep(rel, info) {
			out = append(out, rel)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

func containerPathFromFile(cwd, filePath string) string {
	rel, err := filepath.Rel(cwd, filePath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return filepath.ToSlash(filepath.Base(filePath))
	}
	return filepath.ToSlash(rel)
}

func sanitizeContainerPath(p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" {
		return "", errors.New("container path is empty")
	}
	// Container paths are always slash-separated and must be relative.
	if strings.Contains(p, "\\") {
		return "", fmt.Errorf("invalid container path %q: must use forward slashes", p)
	}
	if strings.HasPrefix(p, "/") {
		return "", fmt.Errorf("invalid container path %q: must be relative", p)
	}
	// Reject Windows drive/UNC-ish prefixes and other schemes.
	if strings.Contains(p, ":") {
		return "", fmt.Errorf("invalid container path %q: must be relative", p)
	}
	// Reject traversal/dot segments before cleaning (don't allow them to be normalized away).
	for _, seg := range strings.Split(p, "/") {
		if seg == "" {
			continue
		}
		if seg == "." || seg == ".." {
			return "", fmt.Errorf("invalid container path %q: path traversal", p)
		}
	}

	clean := path.Clean(p)
	if clean == "." {
		return "", fmt.Errorf("invalid container path %q", p)
	}
	if clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("invalid container path %q: path traversal", p)
	}
	return clean, nil
}

func safeJoinOutput(outDir, containerPath string) (string, error) {
	clean, err := sanitizeContainerPath(containerPath)
	if err != nil {
		return "", err
	}
	dest := filepath.Join(outDir, filepath.FromSlash(clean))
	// Defense-in-depth: ensure dest is within outDir.
	rel, err := filepath.Rel(outDir, dest)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid container path %q: escapes output dir", containerPath)
	}
	return dest, nil
}

func detectMimeType(path string) string {
	m := mimeTypeByExtension(path)
	if m == "" {
		return "application/octet-stream"
	}
	return m
}

func mimeTypeByExtension(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return ""
	}
	return mime.TypeByExtension(ext)
}

func makeIDFromPath(p string) string {
	p = strings.ToLower(p)
	var b strings.Builder
	b.Grow(len(p))
	for _, r := range p {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	id := strings.Trim(b.String(), "_")
	if id == "" {
		return "media"
	}
	return id
}

// validateContainerPaths checks that all markdown paths and media paths/IDs are
// valid container paths before encoding.
func validateContainerPaths(doc *mdocx.Document) error {
	for i, mf := range doc.Markdown.Files {
		if _, err := sanitizeContainerPath(mf.Path); err != nil {
			return fmt.Errorf("markdown file %d: %w", i, err)
		}
	}
	for i, mi := range doc.Media.Items {
		if mi.Path != "" {
			if _, err := sanitizeContainerPath(mi.Path); err != nil {
				return fmt.Errorf("media item %d: %w", i, err)
			}
		}
		if strings.TrimSpace(mi.ID) == "" {
			return fmt.Errorf("media item %d: empty ID", i)
		}
	}
	return nil
}

// humanSize formats a byte count into a human-readable string.
func humanSize(b int) string {
	const (
		kiB = 1024
		miB = 1024 * kiB
		giB = 1024 * miB
	)
	switch {
	case b >= giB:
		return fmt.Sprintf("%.2f GiB", float64(b)/float64(giB))
	case b >= miB:
		return fmt.Sprintf("%.2f MiB", float64(b)/float64(miB))
	case b >= kiB:
		return fmt.Sprintf("%.2f KiB", float64(b)/float64(kiB))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
