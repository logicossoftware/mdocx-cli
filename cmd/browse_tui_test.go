package cmd

import (
	"image"
	"os"
	"testing"

	"github.com/logicossoftware/go-mdocx"
)

func TestScaleImage_NoScaleNeeded(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 80))
	result := scaleImage(img, 200, 200)
	b := result.Bounds()
	if b.Dx() != 100 || b.Dy() != 80 {
		t.Errorf("expected 100x80, got %dx%d", b.Dx(), b.Dy())
	}
}

func TestScaleImage_ScaleDown(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 400, 200))
	result := scaleImage(img, 200, 200)
	b := result.Bounds()
	if b.Dx() != 200 || b.Dy() != 100 {
		t.Errorf("expected 200x100, got %dx%d", b.Dx(), b.Dy())
	}
}

func TestScaleImage_ScaleDownHeight(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 400))
	result := scaleImage(img, 200, 200)
	b := result.Bounds()
	if b.Dx() != 100 || b.Dy() != 200 {
		t.Errorf("expected 100x200, got %dx%d", b.Dx(), b.Dy())
	}
}

func TestScaleImage_ZeroDimensions(t *testing.T) {
	// Test the zero-dimension guard
	img := image.NewRGBA(image.Rect(0, 0, 0, 0))
	result := scaleImage(img, 200, 200)
	// Should return original image unchanged
	if result != img {
		t.Error("expected original image returned for zero dimensions")
	}
}

func TestScaleImage_VeryLargeDownscale(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10000, 5000))
	result := scaleImage(img, 100, 100)
	b := result.Bounds()
	if b.Dx() < 1 || b.Dy() < 1 {
		t.Errorf("scaled image too small: %dx%d", b.Dx(), b.Dy())
	}
	if b.Dx() > 100 || b.Dy() > 100 {
		t.Errorf("scaled image too large: %dx%d", b.Dx(), b.Dy())
	}
}

func TestSupportsSixel_WithWTSession(t *testing.T) {
	// Save original env
	orig := os.Getenv("WT_SESSION")
	origTerm := os.Getenv("TERM")
	origProg := os.Getenv("TERM_PROGRAM")
	defer func() {
		os.Setenv("WT_SESSION", orig)
		os.Setenv("TERM", origTerm)
		os.Setenv("TERM_PROGRAM", origProg)
	}()

	os.Setenv("WT_SESSION", "some-session-id")
	os.Setenv("TERM", "")
	os.Setenv("TERM_PROGRAM", "")
	if !supportsSixel() {
		t.Error("expected supportsSixel() = true with WT_SESSION set")
	}
}

func TestSupportsSixel_WithITerm(t *testing.T) {
	orig := os.Getenv("WT_SESSION")
	origTerm := os.Getenv("TERM")
	origProg := os.Getenv("TERM_PROGRAM")
	defer func() {
		os.Setenv("WT_SESSION", orig)
		os.Setenv("TERM", origTerm)
		os.Setenv("TERM_PROGRAM", origProg)
	}()

	os.Setenv("WT_SESSION", "")
	os.Setenv("TERM", "")
	os.Setenv("TERM_PROGRAM", "iTerm2")
	if !supportsSixel() {
		t.Error("expected supportsSixel() = true with TERM_PROGRAM=iTerm2")
	}
}

func TestSupportsSixel_WithWezTerm(t *testing.T) {
	orig := os.Getenv("WT_SESSION")
	origTerm := os.Getenv("TERM")
	origProg := os.Getenv("TERM_PROGRAM")
	defer func() {
		os.Setenv("WT_SESSION", orig)
		os.Setenv("TERM", origTerm)
		os.Setenv("TERM_PROGRAM", origProg)
	}()

	os.Setenv("WT_SESSION", "")
	os.Setenv("TERM", "")
	os.Setenv("TERM_PROGRAM", "WezTerm")
	if !supportsSixel() {
		t.Error("expected supportsSixel() = true with TERM_PROGRAM=WezTerm")
	}
}

func TestSupportsSixel_WithXterm(t *testing.T) {
	orig := os.Getenv("WT_SESSION")
	origTerm := os.Getenv("TERM")
	origProg := os.Getenv("TERM_PROGRAM")
	defer func() {
		os.Setenv("WT_SESSION", orig)
		os.Setenv("TERM", origTerm)
		os.Setenv("TERM_PROGRAM", origProg)
	}()

	os.Setenv("WT_SESSION", "")
	os.Setenv("TERM_PROGRAM", "")
	os.Setenv("TERM", "xterm-256color")
	if !supportsSixel() {
		t.Error("expected supportsSixel() = true with TERM=xterm-256color")
	}
}

func TestSupportsSixel_NoSixelSupport(t *testing.T) {
	orig := os.Getenv("WT_SESSION")
	origTerm := os.Getenv("TERM")
	origProg := os.Getenv("TERM_PROGRAM")
	defer func() {
		os.Setenv("WT_SESSION", orig)
		os.Setenv("TERM", origTerm)
		os.Setenv("TERM_PROGRAM", origProg)
	}()

	os.Setenv("WT_SESSION", "")
	os.Setenv("TERM_PROGRAM", "")
	os.Setenv("TERM", "")
	if supportsSixel() {
		t.Error("expected supportsSixel() = false with no env vars")
	}
}

func TestSupportsSixel_WithMintty(t *testing.T) {
	orig := os.Getenv("WT_SESSION")
	origTerm := os.Getenv("TERM")
	origProg := os.Getenv("TERM_PROGRAM")
	defer func() {
		os.Setenv("WT_SESSION", orig)
		os.Setenv("TERM", origTerm)
		os.Setenv("TERM_PROGRAM", origProg)
	}()

	os.Setenv("WT_SESSION", "")
	os.Setenv("TERM", "")
	os.Setenv("TERM_PROGRAM", "mintty")
	if !supportsSixel() {
		t.Error("expected supportsSixel() = true with TERM_PROGRAM=mintty")
	}
}

func TestSupportsSixel_WithFoot(t *testing.T) {
	orig := os.Getenv("WT_SESSION")
	origTerm := os.Getenv("TERM")
	origProg := os.Getenv("TERM_PROGRAM")
	defer func() {
		os.Setenv("WT_SESSION", orig)
		os.Setenv("TERM", origTerm)
		os.Setenv("TERM_PROGRAM", origProg)
	}()

	os.Setenv("WT_SESSION", "")
	os.Setenv("TERM_PROGRAM", "")
	os.Setenv("TERM", "foot")
	if !supportsSixel() {
		t.Error("expected supportsSixel() = true with TERM=foot")
	}
}

func TestSupportsSixel_WithContour(t *testing.T) {
	orig := os.Getenv("WT_SESSION")
	origTerm := os.Getenv("TERM")
	origProg := os.Getenv("TERM_PROGRAM")
	defer func() {
		os.Setenv("WT_SESSION", orig)
		os.Setenv("TERM", origTerm)
		os.Setenv("TERM_PROGRAM", origProg)
	}()

	os.Setenv("WT_SESSION", "")
	os.Setenv("TERM", "")
	os.Setenv("TERM_PROGRAM", "contour")
	if !supportsSixel() {
		t.Error("expected supportsSixel() = true with TERM_PROGRAM=contour")
	}
}

func TestBuildRenderer_DefaultTheme(t *testing.T) {
	r, err := buildRenderer("", 80)
	if err != nil {
		t.Fatalf("buildRenderer with empty theme: %v", err)
	}
	if r == nil {
		t.Fatal("expected non-nil renderer")
	}
}

func TestBuildRenderer_CustomTheme(t *testing.T) {
	// "dark" and "light" are built-in glamour styles
	r, err := buildRenderer("dark", 80)
	if err != nil {
		t.Fatalf("buildRenderer with dark theme: %v", err)
	}
	if r == nil {
		t.Fatal("expected non-nil renderer")
	}
}

func TestBuildRenderer_DifferentWidths(t *testing.T) {
	for _, w := range []int{40, 80, 120} {
		r, err := buildRenderer("", w)
		if err != nil {
			t.Fatalf("buildRenderer width=%d: %v", w, err)
		}
		if r == nil {
			t.Fatalf("expected non-nil renderer for width=%d", w)
		}
	}
}

func TestMarkdownItem_Methods(t *testing.T) {
	item := markdownItem{path: "docs/readme.md", size: 1024}

	if got := item.Title(); got != "docs/readme.md" {
		t.Errorf("Title() = %q, want %q", got, "docs/readme.md")
	}
	if got := item.Description(); got != "1024 bytes" {
		t.Errorf("Description() = %q, want %q", got, "1024 bytes")
	}
	if got := item.FilterValue(); got != "docs/readme.md" {
		t.Errorf("FilterValue() = %q, want %q", got, "docs/readme.md")
	}
}

func TestMediaItem_Methods_WithPath(t *testing.T) {
	item := mediaItem{
		id:       "img_png",
		path:     "assets/img.png",
		mimeType: "image/png",
		size:     4096,
	}

	if got := item.Title(); got != "assets/img.png" {
		t.Errorf("Title() = %q, want %q", got, "assets/img.png")
	}
	if got := item.Description(); got != "image/png (4096 bytes)" {
		t.Errorf("Description() = %q, want %q", got, "image/png (4096 bytes)")
	}
	if got := item.FilterValue(); got != "img_png" {
		t.Errorf("FilterValue() = %q, want %q", got, "img_png")
	}
}

func TestMediaItem_Methods_WithoutPath(t *testing.T) {
	item := mediaItem{
		id:       "my_media",
		path:     "",
		mimeType: "application/octet-stream",
		size:     256,
	}

	if got := item.Title(); got != "my_media" {
		t.Errorf("Title() = %q, want %q (should fall back to ID)", got, "my_media")
	}
}

func TestNewBrowseModel_Basic(t *testing.T) {
	doc := &mdocx.Document{
		Metadata: map[string]any{"title": "Test"},
		Markdown: mdocx.MarkdownBundle{
			BundleVersion: 1,
			Files: []mdocx.MarkdownFile{
				{Path: "readme.md", Content: []byte("# Hello")},
			},
		},
		Media: mdocx.MediaBundle{
			BundleVersion: 1,
			Items: []mdocx.MediaItem{
				{ID: "img1", Path: "img.png", MIMEType: "image/png", Data: []byte{1, 2, 3}},
			},
		},
	}

	header := &headerInfo{
		MagicHex:      "4d444f43580d0a1a",
		MagicValid:    true,
		Version:       1,
		FixedHdrSize:  32,
		ReservedClean: true,
	}

	model, err := newBrowseModel(doc, header, "", false)
	if err != nil {
		t.Fatalf("newBrowseModel: %v", err)
	}
	if len(model.tabs) != 4 {
		t.Errorf("expected 4 tabs, got %d", len(model.tabs))
	}
	if model.activeTab != 0 {
		t.Errorf("expected activeTab=0, got %d", model.activeTab)
	}
	if model.metadataView == "(no metadata)" {
		t.Error("expected metadata view to be populated")
	}
	if model.headerView == "(header unavailable)" {
		t.Error("expected header view to be populated")
	}
}

func TestNewBrowseModel_NoMetadata(t *testing.T) {
	doc := &mdocx.Document{
		Markdown: mdocx.MarkdownBundle{BundleVersion: 1},
		Media:    mdocx.MediaBundle{BundleVersion: 1},
	}

	model, err := newBrowseModel(doc, nil, "", false)
	if err != nil {
		t.Fatalf("newBrowseModel: %v", err)
	}
	if model.metadataView != "(no metadata)" {
		t.Errorf("expected '(no metadata)', got %q", model.metadataView)
	}
	if model.headerView != "(header unavailable)" {
		t.Errorf("expected '(header unavailable)', got %q", model.headerView)
	}
}

func TestBrowseModel_Init(t *testing.T) {
	doc := &mdocx.Document{
		Markdown: mdocx.MarkdownBundle{BundleVersion: 1},
		Media:    mdocx.MediaBundle{BundleVersion: 1},
	}

	model, err := newBrowseModel(doc, nil, "", false)
	if err != nil {
		t.Fatal(err)
	}

	cmd := model.Init()
	if cmd == nil {
		t.Error("expected non-nil Init cmd (window title)")
	}
}
