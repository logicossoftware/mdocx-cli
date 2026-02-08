package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/logicossoftware/go-mdocx"
)

func TestWriteUnpacked_OverwriteBlocked(t *testing.T) {
	outDir := t.TempDir()

	doc := &mdocx.Document{
		Metadata: map[string]any{"k": "v"},
		Markdown: mdocx.MarkdownBundle{
			BundleVersion: mdocx.VersionV1,
			Files: []mdocx.MarkdownFile{
				{Path: "readme.md", Content: []byte("content")},
			},
		},
		Media: mdocx.MediaBundle{BundleVersion: mdocx.VersionV1},
	}

	// First unpack should succeed.
	var out bytes.Buffer
	if err := writeUnpacked(doc, outDir, true, &out); err != nil {
		t.Fatalf("first unpack: %v", err)
	}

	// Second unpack WITHOUT force should fail.
	out.Reset()
	err := writeUnpacked(doc, outDir, false, &out)
	if err == nil {
		t.Fatal("expected error for overwrite without --force")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected 'already exists' error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "--force") {
		t.Fatalf("expected '--force' hint in error, got: %v", err)
	}
}

func TestWriteUnpacked_OverwriteAllowedWithForce(t *testing.T) {
	outDir := t.TempDir()

	doc := &mdocx.Document{
		Markdown: mdocx.MarkdownBundle{
			BundleVersion: mdocx.VersionV1,
			Files: []mdocx.MarkdownFile{
				{Path: "readme.md", Content: []byte("original")},
			},
		},
		Media: mdocx.MediaBundle{BundleVersion: mdocx.VersionV1},
	}

	// First unpack.
	var out bytes.Buffer
	if err := writeUnpacked(doc, outDir, true, &out); err != nil {
		t.Fatalf("first unpack: %v", err)
	}

	// Update content and unpack with force.
	doc.Markdown.Files[0].Content = []byte("updated")
	out.Reset()
	if err := writeUnpacked(doc, outDir, true, &out); err != nil {
		t.Fatalf("force unpack: %v", err)
	}

	// Verify content was overwritten.
	data, err := os.ReadFile(filepath.Join(outDir, "readme.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "updated" {
		t.Errorf("expected updated content, got %q", string(data))
	}
}

func TestWriteUnpacked_MediaOverwriteBlocked(t *testing.T) {
	outDir := t.TempDir()

	doc := &mdocx.Document{
		Markdown: mdocx.MarkdownBundle{BundleVersion: mdocx.VersionV1},
		Media: mdocx.MediaBundle{
			BundleVersion: mdocx.VersionV1,
			Items: []mdocx.MediaItem{
				{ID: "img1", Path: "assets/img.png", Data: []byte{1, 2, 3}},
			},
		},
	}

	// First unpack with force.
	var out bytes.Buffer
	if err := writeUnpacked(doc, outDir, true, &out); err != nil {
		t.Fatalf("first unpack: %v", err)
	}

	// Second unpack without force should fail on media.
	out.Reset()
	err := writeUnpacked(doc, outDir, false, &out)
	if err == nil {
		t.Fatal("expected error for media overwrite without --force")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected 'already exists' error, got: %v", err)
	}
}

func TestWriteUnpacked_MediaWithoutPathUsesIDFallback(t *testing.T) {
	outDir := t.TempDir()

	doc := &mdocx.Document{
		Markdown: mdocx.MarkdownBundle{BundleVersion: mdocx.VersionV1},
		Media: mdocx.MediaBundle{
			BundleVersion: mdocx.VersionV1,
			Items: []mdocx.MediaItem{
				{ID: "my_image", Path: "", Data: []byte{1, 2, 3}},
			},
		},
	}

	var out bytes.Buffer
	if err := writeUnpacked(doc, outDir, true, &out); err != nil {
		t.Fatalf("unpack: %v", err)
	}

	// When Path is empty, media should be written to "media/<ID>"
	expected := filepath.Join(outDir, "media", "my_image")
	if _, err := os.Stat(expected); err != nil {
		t.Fatalf("expected media file at %s: %v", expected, err)
	}
}
