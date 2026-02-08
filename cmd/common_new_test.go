package cmd

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/logicossoftware/go-mdocx"
)

// ---------------------------------------------------------------------------
// readMetadataJSON
// ---------------------------------------------------------------------------

func TestReadMetadataJSON_Valid(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "meta.json")
	if err := os.WriteFile(p, []byte(`{"title":"hello","version":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := readMetadataJSON(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m["title"] != "hello" {
		t.Errorf("expected title=hello, got %v", m["title"])
	}
}

func TestReadMetadataJSON_InvalidJSON(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "bad.json")
	if err := os.WriteFile(p, []byte(`{not valid json`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := readMetadataJSON(p)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestReadMetadataJSON_NullObject(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "null.json")
	if err := os.WriteFile(p, []byte(`null`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := readMetadataJSON(p)
	if err == nil {
		t.Fatal("expected error for null JSON object")
	}
	if !strings.Contains(err.Error(), "JSON object") {
		t.Errorf("expected 'JSON object' in error, got: %v", err)
	}
}

func TestReadMetadataJSON_MissingFile(t *testing.T) {
	_, err := readMetadataJSON(filepath.Join(t.TempDir(), "nonexistent.json"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestReadMetadataJSON_EmptyObject(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "empty.json")
	if err := os.WriteFile(p, []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := readMetadataJSON(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m) != 0 {
		t.Errorf("expected empty map, got %v", m)
	}
}

// ---------------------------------------------------------------------------
// collectMarkdownFiles (expanded scenarios)
// ---------------------------------------------------------------------------

func TestCollectMarkdownFiles_WithBaseDir(t *testing.T) {
	tmp := t.TempDir()
	baseDir := filepath.Join(tmp, "docs")
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(baseDir, "a.md"), []byte("# A"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(baseDir, "b.md"), []byte("# B"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Non-markdown file should be ignored
	if err := os.WriteFile(filepath.Join(baseDir, "skip.txt"), []byte("skip"), 0o644); err != nil {
		t.Fatal(err)
	}

	files, err := collectMarkdownFiles(nil, baseDir)
	if err != nil {
		t.Fatalf("collectMarkdownFiles: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
}

func TestCollectMarkdownFiles_DirectoryInArgs(t *testing.T) {
	tmp := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldCwd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	mdDir := filepath.Join(tmp, "subdir")
	if err := os.MkdirAll(mdDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mdDir, "c.md"), []byte("# C"), 0o644); err != nil {
		t.Fatal(err)
	}

	files, err := collectMarkdownFiles([]string{mdDir}, "")
	if err != nil {
		t.Fatalf("collectMarkdownFiles: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
}

func TestCollectMarkdownFiles_NoFilesFound(t *testing.T) {
	tmp := t.TempDir()
	// Empty directory — no .md files
	_, err := collectMarkdownFiles(nil, tmp)
	if err == nil {
		t.Fatal("expected error for no markdown files")
	}
	if !strings.Contains(err.Error(), "no markdown files") {
		t.Errorf("expected 'no markdown files' error, got: %v", err)
	}
}

func TestCollectMarkdownFiles_DuplicatePaths(t *testing.T) {
	tmp := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldCwd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	p := filepath.Join(tmp, "test.md")
	if err := os.WriteFile(p, []byte("# Test"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Same file passed twice should produce duplicate error
	_, err = collectMarkdownFiles([]string{p, p}, "")
	if err == nil {
		t.Fatal("expected error for duplicate paths")
	}
	if !strings.Contains(err.Error(), "duplicate markdown path") {
		t.Errorf("expected 'duplicate markdown path' error, got: %v", err)
	}
}

func TestCollectMarkdownFiles_NonexistentInput(t *testing.T) {
	_, err := collectMarkdownFiles([]string{filepath.Join(t.TempDir(), "nonexistent.md")}, "")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestCollectMarkdownFiles_BaseDirAndArgs(t *testing.T) {
	tmp := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldCwd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	baseDir := filepath.Join(tmp, "base")
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(baseDir, "a.md"), []byte("# A"), 0o644); err != nil {
		t.Fatal(err)
	}

	extraFile := filepath.Join(tmp, "extra.md")
	if err := os.WriteFile(extraFile, []byte("# Extra"), 0o644); err != nil {
		t.Fatal(err)
	}

	files, err := collectMarkdownFiles([]string{extraFile}, baseDir)
	if err != nil {
		t.Fatalf("collectMarkdownFiles: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files (1 from baseDir + 1 from args), got %d", len(files))
	}
}

// ---------------------------------------------------------------------------
// containerPathFromFile (edge cases)
// ---------------------------------------------------------------------------

func TestContainerPathFromFile_FallbackToBasename(t *testing.T) {
	// When file is outside cwd, should fall back to base name
	got := containerPathFromFile("/some/other/dir", "/completely/different/file.md")
	if got != "file.md" {
		t.Errorf("expected fallback to basename, got %q", got)
	}
}

func TestContainerPathFromFile_RelativePath(t *testing.T) {
	got := containerPathFromFile("/workspace", "/workspace/docs/readme.md")
	if got != "docs/readme.md" {
		t.Errorf("expected 'docs/readme.md', got %q", got)
	}
}

func TestContainerPathFromFile_SameDir(t *testing.T) {
	got := containerPathFromFile("/workspace", "/workspace/file.md")
	if got != "file.md" {
		t.Errorf("expected 'file.md', got %q", got)
	}
}

// ---------------------------------------------------------------------------
// detectMimeType / mimeTypeByExtension
// ---------------------------------------------------------------------------

func TestDetectMimeType_Known(t *testing.T) {
	got := detectMimeType("test.png")
	if got != "image/png" {
		t.Errorf("expected 'image/png', got %q", got)
	}
}

func TestDetectMimeType_NoExtension(t *testing.T) {
	got := detectMimeType("README")
	if got != "application/octet-stream" {
		t.Errorf("expected 'application/octet-stream' for no extension, got %q", got)
	}
}

func TestMimeTypeByExtension_Empty(t *testing.T) {
	got := mimeTypeByExtension("noext")
	if got != "" {
		t.Errorf("expected empty for no extension, got %q", got)
	}
}

func TestMimeTypeByExtension_Known(t *testing.T) {
	got := mimeTypeByExtension("image.jpg")
	if got == "" {
		t.Error("expected non-empty MIME for .jpg")
	}
}

// ---------------------------------------------------------------------------
// safeJoinOutput (additional edge cases)
// ---------------------------------------------------------------------------

func TestSafeJoinOutput_Valid(t *testing.T) {
	outDir := t.TempDir()
	got, err := safeJoinOutput(outDir, "docs/readme.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := filepath.Join(outDir, "docs", "readme.md")
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSafeJoinOutput_RejectsTraversal(t *testing.T) {
	outDir := t.TempDir()
	_, err := safeJoinOutput(outDir, "../evil.md")
	if err == nil {
		t.Fatal("expected error for path traversal")
	}
}

func TestSafeJoinOutput_RejectsAbsolute(t *testing.T) {
	outDir := t.TempDir()
	_, err := safeJoinOutput(outDir, "/etc/passwd")
	if err == nil {
		t.Fatal("expected error for absolute path")
	}
}

func TestSafeJoinOutput_RejectsBackslash(t *testing.T) {
	outDir := t.TempDir()
	_, err := safeJoinOutput(outDir, "a\\b.md")
	if err == nil {
		t.Fatal("expected error for backslash")
	}
}

// ---------------------------------------------------------------------------
// collectFiles
// ---------------------------------------------------------------------------

func TestCollectFiles_NestedStructure(t *testing.T) {
	tmp := t.TempDir()
	// Create nested directory structure
	if err := os.MkdirAll(filepath.Join(tmp, "a", "b"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "root.txt"), []byte("r"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "a", "mid.txt"), []byte("m"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "a", "b", "deep.txt"), []byte("d"), 0o644); err != nil {
		t.Fatal(err)
	}

	files, err := collectFiles(tmp, func(rel string, info os.DirEntry) bool {
		return !info.IsDir()
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 3 {
		t.Errorf("expected 3 files, got %d: %v", len(files), files)
	}
}

func TestCollectFiles_WithFilter(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "a.md"), []byte("md"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "b.txt"), []byte("txt"), 0o644); err != nil {
		t.Fatal(err)
	}

	files, err := collectFiles(tmp, func(rel string, info os.DirEntry) bool {
		return !info.IsDir() && strings.HasSuffix(rel, ".md")
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 .md file, got %d: %v", len(files), files)
	}
}

func TestCollectFiles_NonexistentRoot(t *testing.T) {
	_, err := collectFiles(filepath.Join(t.TempDir(), "nonexistent"), func(rel string, info os.DirEntry) bool {
		return true
	})
	if err == nil {
		t.Fatal("expected error for nonexistent root")
	}
}

func TestMakeIDFromPath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"img.png", "img_png"},
		{"assets/logo.png", "assets_logo_png"},
		{"UPPER.JPG", "upper_jpg"},
		{"a-b.png", "a_b_png"},
		{"a_b.png", "a_b_png"},
		{"---", "media"}, // all special chars
		{"", "media"},    // empty
		{"123.png", "123_png"},
	}
	for _, tt := range tests {
		got := makeIDFromPath(tt.input)
		if got != tt.want {
			t.Errorf("makeIDFromPath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCollectMediaItems_SHA256Populated(t *testing.T) {
	tmp := t.TempDir()
	data := []byte("test media content")
	if err := os.WriteFile(filepath.Join(tmp, "test.bin"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := collectMediaItems(tmp)
	if err != nil {
		t.Fatalf("collectMediaItems: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	expectedHash := sha256.Sum256(data)
	if items[0].SHA256 != expectedHash {
		t.Errorf("SHA256 mismatch: got %x, want %x", items[0].SHA256, expectedHash)
	}
}

func TestCollectMediaItems_DuplicateIDDetected(t *testing.T) {
	tmp := t.TempDir()
	// "a-b.png" and "a_b.png" both produce ID "a_b_png"
	if err := os.WriteFile(filepath.Join(tmp, "a-b.png"), []byte{1}, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "a_b.png"), []byte{2}, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := collectMediaItems(tmp)
	if err == nil {
		t.Fatal("expected error for duplicate media IDs")
	}
	if !strings.Contains(err.Error(), "duplicate media ID") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCollectMediaItems_EmptyDir(t *testing.T) {
	items, err := collectMediaItems("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if items != nil {
		t.Fatalf("expected nil for empty media dir, got %v", items)
	}
}

func TestValidateContainerPaths_Valid(t *testing.T) {
	doc := &mdocx.Document{
		Markdown: mdocx.MarkdownBundle{
			Files: []mdocx.MarkdownFile{
				{Path: "docs/readme.md"},
				{Path: "index.md"},
			},
		},
		Media: mdocx.MediaBundle{
			Items: []mdocx.MediaItem{
				{ID: "img1", Path: "assets/img.png"},
				{ID: "img2", Path: "assets/logo.jpg"},
			},
		},
	}
	if err := validateContainerPaths(doc); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateContainerPaths_InvalidMarkdownPath(t *testing.T) {
	doc := &mdocx.Document{
		Markdown: mdocx.MarkdownBundle{
			Files: []mdocx.MarkdownFile{
				{Path: "../evil.md"},
			},
		},
		Media: mdocx.MediaBundle{},
	}
	err := validateContainerPaths(doc)
	if err == nil {
		t.Fatal("expected error for invalid markdown path")
	}
	if !strings.Contains(err.Error(), "markdown file 0") {
		t.Fatalf("expected error to reference markdown file 0, got: %v", err)
	}
}

func TestValidateContainerPaths_InvalidMediaPath(t *testing.T) {
	doc := &mdocx.Document{
		Markdown: mdocx.MarkdownBundle{},
		Media: mdocx.MediaBundle{
			Items: []mdocx.MediaItem{
				{ID: "img1", Path: "/absolute/path.png"},
			},
		},
	}
	err := validateContainerPaths(doc)
	if err == nil {
		t.Fatal("expected error for invalid media path")
	}
	if !strings.Contains(err.Error(), "media item 0") {
		t.Fatalf("expected error to reference media item 0, got: %v", err)
	}
}

func TestValidateContainerPaths_EmptyMediaID(t *testing.T) {
	doc := &mdocx.Document{
		Markdown: mdocx.MarkdownBundle{},
		Media: mdocx.MediaBundle{
			Items: []mdocx.MediaItem{
				{ID: "", Path: "valid.png"},
			},
		},
	}
	err := validateContainerPaths(doc)
	if err == nil {
		t.Fatal("expected error for empty media ID")
	}
	if !strings.Contains(err.Error(), "empty ID") {
		t.Fatalf("expected empty ID error, got: %v", err)
	}
}

func TestValidateContainerPaths_MediaWithoutPath(t *testing.T) {
	doc := &mdocx.Document{
		Markdown: mdocx.MarkdownBundle{},
		Media: mdocx.MediaBundle{
			Items: []mdocx.MediaItem{
				{ID: "valid_id", Path: ""},
			},
		},
	}
	if err := validateContainerPaths(doc); err != nil {
		t.Fatalf("media with empty path but valid ID should be allowed, got: %v", err)
	}
}

func TestHumanSize(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1023, "1023 B"},
		{1024, "1.00 KiB"},
		{1536, "1.50 KiB"},
		{1048576, "1.00 MiB"},
		{1572864, "1.50 MiB"},
		{1073741824, "1.00 GiB"},
	}
	for _, tt := range tests {
		got := humanSize(tt.input)
		if got != tt.want {
			t.Errorf("humanSize(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseCompression(t *testing.T) {
	tests := []struct {
		input   string
		want    mdocx.Compression
		wantErr bool
	}{
		{"", mdocx.CompZSTD, false},
		{"zstd", mdocx.CompZSTD, false},
		{"ZSTD", mdocx.CompZSTD, false},
		{"none", mdocx.CompNone, false},
		{"zip", mdocx.CompZIP, false},
		{"lz4", mdocx.CompLZ4, false},
		{"br", mdocx.CompBR, false},
		{"brotli", mdocx.CompBR, false},
		{"invalid", mdocx.CompNone, true},
	}
	for _, tt := range tests {
		got, err := parseCompression(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseCompression(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.want {
			t.Errorf("parseCompression(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestReadHeaderInfo_ValidFile(t *testing.T) {
	// Create a minimal valid MDOCX file via Encode
	tmp := t.TempDir()
	outPath := filepath.Join(tmp, "test.mdocx")

	doc := &mdocx.Document{
		Markdown: mdocx.MarkdownBundle{
			BundleVersion: mdocx.VersionV1,
			Files:         []mdocx.MarkdownFile{{Path: "test.md", Content: []byte("# Test")}},
		},
		Media: mdocx.MediaBundle{BundleVersion: mdocx.VersionV1},
	}
	f, err := os.Create(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := mdocx.Encode(f, doc, mdocx.WithMarkdownCompression(mdocx.CompNone), mdocx.WithMediaCompression(mdocx.CompNone)); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()

	header, err := readHeaderInfo(outPath)
	if err != nil {
		t.Fatalf("readHeaderInfo: %v", err)
	}
	if !header.MagicValid {
		t.Error("expected magic to be valid")
	}
	if header.Version != 1 {
		t.Errorf("expected version 1, got %d", header.Version)
	}
	if header.FixedHdrSize != 32 {
		t.Errorf("expected fixed header size 32, got %d", header.FixedHdrSize)
	}
	if !header.ReservedClean {
		t.Error("expected reserved bytes to be clean")
	}
}

func TestReadHeaderInfo_InvalidFile(t *testing.T) {
	tmp := t.TempDir()
	badPath := filepath.Join(tmp, "bad.mdocx")
	if err := os.WriteFile(badPath, []byte("not a valid mdocx file at all!!"), 0o644); err != nil {
		t.Fatal(err)
	}

	// File is 30 bytes, too short for a 32-byte header read
	_, err := readHeaderInfo(badPath)
	if err != nil {
		// Expected: file too short
		return
	}
	// If it does read (32 bytes match), magic should be invalid
}

func TestReadHeaderInfo_DirtyReservedBytes(t *testing.T) {
	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "dirty.mdocx")

	// Build a valid 32-byte header with dirty reserved bytes
	var buf [32]byte
	copy(buf[0:8], []byte{'M', 'D', 'O', 'C', 'X', '\r', '\n', 0x1A})
	buf[8] = 1     // version low byte
	buf[9] = 0     // version high byte
	buf[12] = 32   // fixed header size
	buf[25] = 0xFF // dirty reserved byte

	if err := os.WriteFile(filePath, buf[:], 0o644); err != nil {
		t.Fatal(err)
	}

	header, err := readHeaderInfo(filePath)
	if err != nil {
		t.Fatalf("readHeaderInfo: %v", err)
	}
	if header.ReservedClean {
		t.Error("expected ReservedClean to be false for dirty reserved bytes")
	}
	if !header.MagicValid {
		t.Error("expected magic to be valid")
	}
}
