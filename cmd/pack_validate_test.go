package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/logicossoftware/go-mdocx"
)

func TestPackHelpers_RoundTripEncodeDecode(t *testing.T) {
	tmp := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() { _ = os.Chdir(oldCwd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	mdPath := filepath.Join(tmp, "docs", "readme.md")
	if err := os.MkdirAll(filepath.Dir(mdPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(mdPath, []byte("# Hello\n"), 0o644); err != nil {
		t.Fatalf("write md: %v", err)
	}

	mediaDir := filepath.Join(tmp, "media")
	if err := os.MkdirAll(mediaDir, 0o755); err != nil {
		t.Fatalf("mkdir media: %v", err)
	}
	mediaPath := filepath.Join(mediaDir, "img.png")
	if err := os.WriteFile(mediaPath, []byte{1, 2, 3, 4}, 0o644); err != nil {
		t.Fatalf("write media: %v", err)
	}

	mdFiles, err := collectMarkdownFiles([]string{mdPath}, "")
	if err != nil {
		t.Fatalf("collectMarkdownFiles: %v", err)
	}
	if len(mdFiles) != 1 || mdFiles[0].Path != "docs/readme.md" {
		t.Fatalf("unexpected markdown file path: %#v", mdFiles)
	}

	mediaItems, err := collectMediaItems(mediaDir)
	if err != nil {
		t.Fatalf("collectMediaItems: %v", err)
	}
	if len(mediaItems) != 1 {
		t.Fatalf("expected 1 media item, got %d", len(mediaItems))
	}
	if mediaItems[0].Path != "img.png" {
		t.Fatalf("unexpected media path: %q", mediaItems[0].Path)
	}
	if mediaItems[0].ID != "img_png" {
		t.Fatalf("unexpected media id: %q", mediaItems[0].ID)
	}

	doc := &mdocx.Document{
		Metadata: map[string]any{"k": "v"},
		Markdown: mdocx.MarkdownBundle{BundleVersion: mdocx.VersionV1, RootPath: "", Files: mdFiles},
		Media:    mdocx.MediaBundle{BundleVersion: mdocx.VersionV1, Items: mediaItems},
	}

	var buf bytes.Buffer
	if err := mdocx.Encode(&buf, doc, mdocx.WithMarkdownCompression(mdocx.CompNone), mdocx.WithMediaCompression(mdocx.CompNone)); err != nil {
		t.Fatalf("encode: %v", err)
	}

	decoded, err := mdocx.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	result := buildValidationResult(decoded, nil, nil)
	if !result.Valid {
		t.Fatalf("expected valid")
	}
	if result.MarkdownFileCount != 1 || result.MediaItemCount != 1 {
		t.Fatalf("unexpected counts: %#v", result)
	}
	if string(decoded.Markdown.Files[0].Content) != "# Hello\n" {
		t.Fatalf("markdown content mismatch")
	}
	if len(decoded.Media.Items[0].Data) != 4 {
		t.Fatalf("media content mismatch")
	}
}
