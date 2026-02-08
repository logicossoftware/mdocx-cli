package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/logicossoftware/go-mdocx"
	"github.com/spf13/cobra"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// createTestMDOCX writes a valid .mdocx file to outPath and returns the path.
func createTestMDOCX(t *testing.T, outPath string, meta map[string]any) string {
	t.Helper()
	doc := &mdocx.Document{
		Metadata: meta,
		Markdown: mdocx.MarkdownBundle{
			BundleVersion: mdocx.VersionV1,
			RootPath:      "readme.md",
			Files: []mdocx.MarkdownFile{
				{Path: "readme.md", Content: []byte("# Hello World\n\nThis is a test.")},
				{Path: "docs/guide.md", Content: []byte("# Guide\n\nSome content.")},
			},
		},
		Media: mdocx.MediaBundle{
			BundleVersion: mdocx.VersionV1,
			Items: []mdocx.MediaItem{
				{ID: "logo_png", Path: "assets/logo.png", MIMEType: "image/png", Data: []byte{0x89, 'P', 'N', 'G'}},
			},
		},
	}
	f, err := os.Create(outPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := mdocx.Encode(f, doc,
		mdocx.WithMarkdownCompression(mdocx.CompNone),
		mdocx.WithMediaCompression(mdocx.CompNone),
	); err != nil {
		t.Fatal(err)
	}
	return outPath
}

// executeCommand runs a cobra command with args, capturing stdout.
func executeCommand(root *cobra.Command, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

// ---------------------------------------------------------------------------
// version command
// ---------------------------------------------------------------------------

func TestVersionCommand(t *testing.T) {
	output, err := executeCommand(rootCmd, "version")
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}
	if !strings.Contains(output, "mdocx") {
		t.Errorf("expected 'mdocx' in output, got: %s", output)
	}
	if !strings.Contains(output, "Commit:") {
		t.Errorf("expected 'Commit:' in output, got: %s", output)
	}
	if !strings.Contains(output, "Built:") {
		t.Errorf("expected 'Built:' in output, got: %s", output)
	}
}

// ---------------------------------------------------------------------------
// inspect command
// ---------------------------------------------------------------------------

func TestInspectCommand_Text(t *testing.T) {
	tmp := t.TempDir()
	mdocxPath := createTestMDOCX(t, filepath.Join(tmp, "test.mdocx"), map[string]any{"title": "Test"})

	output, err := executeCommand(rootCmd, "inspect", mdocxPath)
	if err != nil {
		t.Fatalf("inspect command failed: %v", err)
	}
	if !strings.Contains(output, "Markdown files") {
		t.Errorf("expected 'Markdown files' in output, got: %s", output)
	}
	if !strings.Contains(output, "Media IDs") {
		t.Errorf("expected 'Media IDs' in output, got: %s", output)
	}
	if !strings.Contains(output, "Root path:") {
		t.Errorf("expected 'Root path:' in output, got: %s", output)
	}
	if !strings.Contains(output, "Bundle versions:") {
		t.Errorf("expected 'Bundle versions:' in output, got: %s", output)
	}
}

func TestInspectCommand_JSON(t *testing.T) {
	tmp := t.TempDir()
	mdocxPath := createTestMDOCX(t, filepath.Join(tmp, "test.mdocx"), map[string]any{"title": "Test"})

	output, err := executeCommand(rootCmd, "inspect", "--json", mdocxPath)
	if err != nil {
		t.Fatalf("inspect --json command failed: %v", err)
	}

	var summary inspectSummary
	if err := json.Unmarshal([]byte(output), &summary); err != nil {
		t.Fatalf("failed to parse JSON output: %v\noutput: %s", err, output)
	}
	if len(summary.MarkdownFiles) != 2 {
		t.Errorf("expected 2 markdown files, got %d", len(summary.MarkdownFiles))
	}
	if len(summary.MediaIDs) != 1 {
		t.Errorf("expected 1 media ID, got %d", len(summary.MediaIDs))
	}
	if summary.RootPath != "readme.md" {
		t.Errorf("expected root_path 'readme.md', got %q", summary.RootPath)
	}
	if summary.MarkdownBundleVersion != 1 {
		t.Errorf("expected markdown bundle version 1, got %d", summary.MarkdownBundleVersion)
	}
}

func TestInspectCommand_NoMetadata(t *testing.T) {
	tmp := t.TempDir()
	mdocxPath := createTestMDOCX(t, filepath.Join(tmp, "test.mdocx"), nil)

	output, err := executeCommand(rootCmd, "inspect", "--json", mdocxPath)
	if err != nil {
		t.Fatalf("inspect --json failed: %v", err)
	}

	var summary inspectSummary
	if err := json.Unmarshal([]byte(output), &summary); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	// metadata_keys should be [] not null
	if summary.MetadataKeys == nil {
		t.Error("expected non-nil MetadataKeys")
	}
}

func TestInspectCommand_MissingFile(t *testing.T) {
	_, err := executeCommand(rootCmd, "inspect", filepath.Join(t.TempDir(), "nonexistent.mdocx"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

// ---------------------------------------------------------------------------
// validate command
// ---------------------------------------------------------------------------

func TestValidateCommand_Valid(t *testing.T) {
	tmp := t.TempDir()
	mdocxPath := createTestMDOCX(t, filepath.Join(tmp, "test.mdocx"), nil)

	output, err := executeCommand(rootCmd, "validate", mdocxPath)
	if err != nil {
		t.Fatalf("validate command failed: %v", err)
	}
	if !strings.Contains(output, "Valid MDOCX") {
		t.Errorf("expected 'Valid MDOCX' in output, got: %s", output)
	}
}

func TestValidateCommand_ValidJSON(t *testing.T) {
	tmp := t.TempDir()
	mdocxPath := createTestMDOCX(t, filepath.Join(tmp, "test.mdocx"), nil)

	output, err := executeCommand(rootCmd, "validate", "--json", mdocxPath)
	if err != nil {
		t.Fatalf("validate --json failed: %v", err)
	}

	var result validationResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid, got warnings: %v", result.Warnings)
	}
	if result.MarkdownFileCount != 2 {
		t.Errorf("expected 2 markdown files, got %d", result.MarkdownFileCount)
	}
	if result.MediaItemCount != 1 {
		t.Errorf("expected 1 media item, got %d", result.MediaItemCount)
	}
}

func TestValidateCommand_MissingFile(t *testing.T) {
	_, err := executeCommand(rootCmd, "validate", filepath.Join(t.TempDir(), "nonexistent.mdocx"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestValidateCommand_CorruptFile(t *testing.T) {
	tmp := t.TempDir()
	badPath := filepath.Join(tmp, "bad.mdocx")
	if err := os.WriteFile(badPath, []byte("this is not a valid mdocx file"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := executeCommand(rootCmd, "validate", badPath)
	if err == nil {
		t.Fatal("expected error for corrupt file")
	}
}

// ---------------------------------------------------------------------------
// unpack command
// ---------------------------------------------------------------------------

func TestUnpackCommand(t *testing.T) {
	tmp := t.TempDir()
	mdocxPath := createTestMDOCX(t, filepath.Join(tmp, "test.mdocx"), map[string]any{"k": "v"})
	outDir := filepath.Join(tmp, "unpacked")

	output, err := executeCommand(rootCmd, "unpack", "-o", outDir, "--force", mdocxPath)
	if err != nil {
		t.Fatalf("unpack command failed: %v\noutput: %s", err, output)
	}

	// Check files were written
	if _, err := os.Stat(filepath.Join(outDir, "readme.md")); err != nil {
		t.Error("expected readme.md to be written")
	}
	if _, err := os.Stat(filepath.Join(outDir, "docs", "guide.md")); err != nil {
		t.Error("expected docs/guide.md to be written")
	}
	if _, err := os.Stat(filepath.Join(outDir, "assets", "logo.png")); err != nil {
		t.Error("expected assets/logo.png to be written")
	}
	if _, err := os.Stat(filepath.Join(outDir, "metadata.json")); err != nil {
		t.Error("expected metadata.json to be written")
	}
}

func TestUnpackCommand_NoForce_BlocksOverwrite(t *testing.T) {
	tmp := t.TempDir()
	mdocxPath := createTestMDOCX(t, filepath.Join(tmp, "test.mdocx"), nil)
	outDir := filepath.Join(tmp, "unpacked")

	// First unpack with --force
	_, err := executeCommand(rootCmd, "unpack", "-o", outDir, "--force=true", mdocxPath)
	if err != nil {
		t.Fatalf("first unpack failed: %v", err)
	}

	// Second unpack without --force should fail
	_, err = executeCommand(rootCmd, "unpack", "-o", outDir, "--force=false", mdocxPath)
	if err == nil {
		t.Fatal("expected error for overwrite without --force")
	}
}

func TestUnpackCommand_MissingFile(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "out")
	_, err := executeCommand(rootCmd, "unpack", "-o", outDir, filepath.Join(t.TempDir(), "nonexistent.mdocx"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

// ---------------------------------------------------------------------------
// pack command
// ---------------------------------------------------------------------------

func TestPackCommand_Basic(t *testing.T) {
	tmp := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldCwd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	// Create a markdown file
	mdFile := filepath.Join(tmp, "hello.md")
	if err := os.WriteFile(mdFile, []byte("# Hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	outPath := filepath.Join(tmp, "output.mdocx")
	output, err := executeCommand(rootCmd, "pack", "--compression", "none", "--markdown-dir", "", "--media-dir", "", "--metadata", "", "--root", "", "-o", outPath, mdFile)
	if err != nil {
		t.Fatalf("pack command failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "Packed") {
		t.Errorf("expected 'Packed' in output, got: %s", output)
	}

	// Verify the file exists and is valid
	if _, err := os.Stat(outPath); err != nil {
		t.Fatal("expected output file to exist")
	}

	// Validate the created file
	_, err = executeCommand(rootCmd, "validate", outPath)
	if err != nil {
		t.Fatalf("validation of packed file failed: %v", err)
	}
}

func TestPackCommand_WithMedia(t *testing.T) {
	tmp := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldCwd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	// Create markdown
	mdFile := filepath.Join(tmp, "index.md")
	if err := os.WriteFile(mdFile, []byte("# Index"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create media
	mediaDir := filepath.Join(tmp, "media")
	if err := os.MkdirAll(mediaDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mediaDir, "image.png"), []byte{1, 2, 3, 4}, 0o644); err != nil {
		t.Fatal(err)
	}

	outPath := filepath.Join(tmp, "output.mdocx")
	output, err := executeCommand(rootCmd, "pack",
		"--compression", "none",
		"--media-dir", mediaDir,
		"--root", "index.md",
		"--metadata", "",
		"-o", outPath,
		mdFile,
	)
	if err != nil {
		t.Fatalf("pack with media failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "1 media items") {
		t.Errorf("expected '1 media items' in output, got: %s", output)
	}
}

func TestPackCommand_WithMetadata(t *testing.T) {
	tmp := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldCwd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	mdFile := filepath.Join(tmp, "doc.md")
	if err := os.WriteFile(mdFile, []byte("# Doc"), 0o644); err != nil {
		t.Fatal(err)
	}

	metaFile := filepath.Join(tmp, "meta.json")
	if err := os.WriteFile(metaFile, []byte(`{"title":"My Bundle","version":1}`), 0o644); err != nil {
		t.Fatal(err)
	}

	outPath := filepath.Join(tmp, "output.mdocx")
	output, err := executeCommand(rootCmd, "pack",
		"--compression", "none",
		"--metadata", metaFile,
		"--root", "",
		"--media-dir", "",
		"-o", outPath,
		mdFile,
	)
	if err != nil {
		t.Fatalf("pack with metadata failed: %v\noutput: %s", err, output)
	}

	// Inspect to verify metadata
	inspOut, err := executeCommand(rootCmd, "inspect", "--json", outPath)
	if err != nil {
		t.Fatalf("inspect failed: %v", err)
	}
	var summary inspectSummary
	if err := json.Unmarshal([]byte(inspOut), &summary); err != nil {
		t.Fatalf("parse inspect JSON: %v", err)
	}
	if len(summary.MetadataKeys) == 0 {
		t.Error("expected metadata keys in output")
	}
}

func TestPackCommand_NoInputs(t *testing.T) {
	_, err := executeCommand(rootCmd, "pack", "--compression", "none")
	if err == nil {
		t.Fatal("expected error when no inputs provided")
	}
}

func TestPackCommand_WithMarkdownDir(t *testing.T) {
	tmp := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldCwd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	mdDir := filepath.Join(tmp, "docs")
	if err := os.MkdirAll(mdDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mdDir, "a.md"), []byte("# A"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mdDir, "b.md"), []byte("# B"), 0o644); err != nil {
		t.Fatal(err)
	}

	outPath := filepath.Join(tmp, "output.mdocx")
	output, err := executeCommand(rootCmd, "pack",
		"--compression", "none",
		"--markdown-dir", mdDir,
		"--root", "",
		"--media-dir", "",
		"--metadata", "",
		"-o", outPath,
	)
	if err != nil {
		t.Fatalf("pack with --markdown-dir failed: %v\noutput: %s", err, output)
	}
	if !strings.Contains(output, "2 markdown files") {
		t.Errorf("expected '2 markdown files' in output, got: %s", output)
	}
}

func TestPackCommand_InvalidCompression(t *testing.T) {
	tmp := t.TempDir()
	mdFile := filepath.Join(tmp, "test.md")
	if err := os.WriteFile(mdFile, []byte("# Test"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := executeCommand(rootCmd, "pack",
		"--compression", "invalid_comp",
		"-o", filepath.Join(tmp, "out.mdocx"),
		mdFile,
	)
	if err == nil {
		t.Fatal("expected error for invalid compression")
	}
}

func TestPackCommand_InvalidRootPath(t *testing.T) {
	tmp := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldCwd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	mdFile := filepath.Join(tmp, "test.md")
	if err := os.WriteFile(mdFile, []byte("# Test"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err = executeCommand(rootCmd, "pack",
		"--compression", "none",
		"--root", "nonexistent.md",
		"-o", filepath.Join(tmp, "out.mdocx"),
		mdFile,
	)
	if err == nil {
		t.Fatal("expected error for invalid root path")
	}
}

// ---------------------------------------------------------------------------
// pack -> unpack roundtrip
// ---------------------------------------------------------------------------

func TestPackUnpackRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldCwd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	// Create source files
	if err := os.WriteFile(filepath.Join(tmp, "readme.md"), []byte("# Readme\nContent here."), 0o644); err != nil {
		t.Fatal(err)
	}
	mediaDir := filepath.Join(tmp, "media")
	if err := os.MkdirAll(mediaDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mediaDir, "data.bin"), []byte{10, 20, 30}, 0o644); err != nil {
		t.Fatal(err)
	}
	metaPath := filepath.Join(tmp, "meta.json")
	if err := os.WriteFile(metaPath, []byte(`{"author":"test"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Pack
	bundlePath := filepath.Join(tmp, "bundle.mdocx")
	_, err = executeCommand(rootCmd, "pack",
		"--compression", "none",
		"--media-dir", mediaDir,
		"--metadata", metaPath,
		"--markdown-dir", "",
		"--root", "",
		"-o", bundlePath,
		filepath.Join(tmp, "readme.md"),
	)
	if err != nil {
		t.Fatalf("pack: %v", err)
	}

	// Unpack
	outDir := filepath.Join(tmp, "unpacked")
	_, err = executeCommand(rootCmd, "unpack", "-o", outDir, "--force=true", bundlePath)
	if err != nil {
		t.Fatalf("unpack: %v", err)
	}

	// Verify unpacked content
	mdData, err := os.ReadFile(filepath.Join(outDir, "readme.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(mdData) != "# Readme\nContent here." {
		t.Errorf("markdown content mismatch: %q", string(mdData))
	}

	mediaData, err := os.ReadFile(filepath.Join(outDir, "data.bin"))
	if err != nil {
		t.Fatal(err)
	}
	if len(mediaData) != 3 || mediaData[0] != 10 {
		t.Errorf("media content mismatch: %v", mediaData)
	}

	if _, err := os.Stat(filepath.Join(outDir, "metadata.json")); err != nil {
		t.Error("expected metadata.json")
	}
}
