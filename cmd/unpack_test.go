package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/logicossoftware/go-mdocx"
)

func TestWriteUnpacked_BlocksPathTraversal(t *testing.T) {
	outDir := t.TempDir()

	doc := &mdocx.Document{
		Markdown: mdocx.MarkdownBundle{
			BundleVersion: mdocx.VersionV1,
			Files: []mdocx.MarkdownFile{
				{Path: "../evil.md", Content: []byte("nope")},
			},
		},
		Media: mdocx.MediaBundle{BundleVersion: mdocx.VersionV1},
	}

	var out bytes.Buffer
	err := writeUnpacked(doc, outDir, &out)
	if err == nil {
		t.Fatalf("expected error")
	}

	// Also block traversal via embedded segments.
	doc.Markdown.Files[0].Path = "a/../evil.md"
	err = writeUnpacked(doc, outDir, &out)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestWriteUnpacked_WritesWithinOutputDir(t *testing.T) {
	outDir := t.TempDir()

	doc := &mdocx.Document{
		Metadata: map[string]any{"k": "v"},
		Markdown: mdocx.MarkdownBundle{
			BundleVersion: mdocx.VersionV1,
			Files: []mdocx.MarkdownFile{
				{Path: "docs/readme.md", Content: []byte("hello")},
			},
		},
		Media: mdocx.MediaBundle{
			BundleVersion: mdocx.VersionV1,
			Items:         []mdocx.MediaItem{{ID: "img1", Path: "media/img.png", Data: []byte{1, 2, 3}}},
		},
	}

	var out bytes.Buffer
	if err := writeUnpacked(doc, outDir, &out); err != nil {
		t.Fatalf("writeUnpacked: %v", err)
	}

	wantMd := filepath.Join(outDir, "docs", "readme.md")
	if _, err := os.Stat(wantMd); err != nil {
		t.Fatalf("expected markdown file written at %s: %v", wantMd, err)
	}

	wantMedia := filepath.Join(outDir, "media", "img.png")
	if _, err := os.Stat(wantMedia); err != nil {
		t.Fatalf("expected media file written at %s: %v", wantMedia, err)
	}

	wantMeta := filepath.Join(outDir, "metadata.json")
	if _, err := os.Stat(wantMeta); err != nil {
		t.Fatalf("expected metadata.json written at %s: %v", wantMeta, err)
	}
}
