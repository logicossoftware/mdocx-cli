package cmd

import (
	"encoding/json"
	"testing"

	"github.com/logicossoftware/go-mdocx"
)

func TestBuildValidationResult_ValidDoc(t *testing.T) {
	doc := &mdocx.Document{
		Markdown: mdocx.MarkdownBundle{
			BundleVersion: 1,
			Files: []mdocx.MarkdownFile{
				{Path: "a.md", Content: []byte("hello")},
				{Path: "b.md", Content: []byte("world")},
			},
		},
		Media: mdocx.MediaBundle{
			BundleVersion: 1,
			Items: []mdocx.MediaItem{
				{ID: "img1", Data: []byte{1, 2, 3}},
			},
		},
	}

	header := &headerInfo{
		MagicValid:    true,
		Version:       1,
		FixedHdrSize:  32,
		ReservedClean: true,
	}

	result := buildValidationResult(doc, header, nil)
	if !result.Valid {
		t.Fatalf("expected valid, got warnings: %v", result.Warnings)
	}
	if result.MarkdownFileCount != 2 {
		t.Errorf("expected 2 markdown files, got %d", result.MarkdownFileCount)
	}
	if result.MediaItemCount != 1 {
		t.Errorf("expected 1 media item, got %d", result.MediaItemCount)
	}
	if result.TotalMarkdownBytes != 10 {
		t.Errorf("expected 10 total markdown bytes, got %d", result.TotalMarkdownBytes)
	}
	if result.TotalMediaBytes != 3 {
		t.Errorf("expected 3 total media bytes, got %d", result.TotalMediaBytes)
	}
}

func TestBuildValidationResult_BadBundleVersion(t *testing.T) {
	doc := &mdocx.Document{
		Markdown: mdocx.MarkdownBundle{BundleVersion: 99},
		Media:    mdocx.MediaBundle{BundleVersion: 1},
	}

	result := buildValidationResult(doc, nil, nil)
	if result.Valid {
		t.Fatal("expected invalid for bad bundle version")
	}
	found := false
	for _, w := range result.Warnings {
		if w == "markdown BundleVersion is 99, expected 1" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected warning about markdown BundleVersion, got: %v", result.Warnings)
	}
}

func TestBuildValidationResult_DuplicateMarkdownPaths(t *testing.T) {
	doc := &mdocx.Document{
		Markdown: mdocx.MarkdownBundle{
			BundleVersion: 1,
			Files: []mdocx.MarkdownFile{
				{Path: "same.md"},
				{Path: "same.md"},
			},
		},
		Media: mdocx.MediaBundle{BundleVersion: 1},
	}

	result := buildValidationResult(doc, nil, nil)
	if result.Valid {
		t.Fatal("expected invalid for duplicate markdown paths")
	}
	found := false
	for _, w := range result.Warnings {
		if w == `duplicate markdown path: "same.md"` {
			found = true
		}
	}
	if !found {
		t.Errorf("expected duplicate path warning, got: %v", result.Warnings)
	}
}

func TestBuildValidationResult_DuplicateMediaIDs(t *testing.T) {
	doc := &mdocx.Document{
		Markdown: mdocx.MarkdownBundle{BundleVersion: 1},
		Media: mdocx.MediaBundle{
			BundleVersion: 1,
			Items: []mdocx.MediaItem{
				{ID: "dup"},
				{ID: "dup"},
			},
		},
	}

	result := buildValidationResult(doc, nil, nil)
	if result.Valid {
		t.Fatal("expected invalid for duplicate media IDs")
	}
	found := false
	for _, w := range result.Warnings {
		if w == `duplicate media ID: "dup"` {
			found = true
		}
	}
	if !found {
		t.Errorf("expected duplicate ID warning, got: %v", result.Warnings)
	}
}

func TestBuildValidationResult_HeaderChecks(t *testing.T) {
	doc := &mdocx.Document{
		Markdown: mdocx.MarkdownBundle{BundleVersion: 1},
		Media:    mdocx.MediaBundle{BundleVersion: 1},
	}

	tests := []struct {
		name    string
		header  *headerInfo
		wantMsg string
	}{
		{
			name:    "invalid magic",
			header:  &headerInfo{MagicValid: false, Version: 1, FixedHdrSize: 32, ReservedClean: true},
			wantMsg: "invalid magic bytes",
		},
		{
			name:    "wrong version",
			header:  &headerInfo{MagicValid: true, Version: 2, FixedHdrSize: 32, ReservedClean: true},
			wantMsg: "header version is 2, expected 1",
		},
		{
			name:    "wrong header size",
			header:  &headerInfo{MagicValid: true, Version: 1, FixedHdrSize: 64, ReservedClean: true},
			wantMsg: "fixed header size is 64, expected 32",
		},
		{
			name:    "dirty reserved",
			header:  &headerInfo{MagicValid: true, Version: 1, FixedHdrSize: 32, ReservedClean: false},
			wantMsg: "reserved header bytes are not zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildValidationResult(doc, tt.header, nil)
			if result.Valid {
				t.Fatal("expected invalid")
			}
			found := false
			for _, w := range result.Warnings {
				if w == tt.wantMsg {
					found = true
				}
			}
			if !found {
				t.Errorf("expected warning %q, got: %v", tt.wantMsg, result.Warnings)
			}
		})
	}
}

func TestBuildValidationResult_HeaderError(t *testing.T) {
	doc := &mdocx.Document{
		Markdown: mdocx.MarkdownBundle{BundleVersion: 1},
		Media:    mdocx.MediaBundle{BundleVersion: 1},
	}

	result := buildValidationResult(doc, nil, errForTest("header broke"))
	if result.Valid {
		t.Fatal("expected invalid when header has error")
	}
	found := false
	for _, w := range result.Warnings {
		if w == "header read error: header broke" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected header read error warning, got: %v", result.Warnings)
	}
}

// errForTest is a simple error type for testing.
type errForTest string

func (e errForTest) Error() string { return string(e) }

func TestBuildInspectSummary_EmptySlicesNotNull(t *testing.T) {
	doc := &mdocx.Document{
		Markdown: mdocx.MarkdownBundle{BundleVersion: 1},
		Media:    mdocx.MediaBundle{BundleVersion: 1},
	}

	summary := buildInspectSummary(doc, nil)

	// Verify slices are non-nil (serialize as [] not null)
	b, _ := json.Marshal(summary)
	s := string(b)

	if !jsonContains(s, `"metadata_keys":[]`) {
		t.Error("expected metadata_keys to be [], got null or missing")
	}
	if !jsonContains(s, `"markdown_files":[]`) {
		t.Error("expected markdown_files to be [], got null or missing")
	}
	if !jsonContains(s, `"media_ids":[]`) {
		t.Error("expected media_ids to be [], got null or missing")
	}
	if !jsonContains(s, `"media_paths":[]`) {
		t.Error("expected media_paths to be [], got null or missing")
	}
}

func TestBuildInspectSummary_IncludesBundleVersionsAndRootPath(t *testing.T) {
	doc := &mdocx.Document{
		Markdown: mdocx.MarkdownBundle{
			BundleVersion: 1,
			RootPath:      "index.md",
			Files: []mdocx.MarkdownFile{
				{Path: "index.md", Content: []byte("# Index")},
			},
		},
		Media: mdocx.MediaBundle{BundleVersion: 1},
	}

	summary := buildInspectSummary(doc, nil)
	if summary.MarkdownBundleVersion != 1 {
		t.Errorf("expected MarkdownBundleVersion 1, got %d", summary.MarkdownBundleVersion)
	}
	if summary.MediaBundleVersion != 1 {
		t.Errorf("expected MediaBundleVersion 1, got %d", summary.MediaBundleVersion)
	}
	if summary.RootPath != "index.md" {
		t.Errorf("expected RootPath %q, got %q", "index.md", summary.RootPath)
	}
}

func TestBuildInspectSummary_TotalSizes(t *testing.T) {
	doc := &mdocx.Document{
		Markdown: mdocx.MarkdownBundle{
			BundleVersion: 1,
			Files: []mdocx.MarkdownFile{
				{Path: "a.md", Content: []byte("hello")},   // 5 bytes
				{Path: "b.md", Content: []byte("world!!")}, // 7 bytes
			},
		},
		Media: mdocx.MediaBundle{
			BundleVersion: 1,
			Items: []mdocx.MediaItem{
				{ID: "img1", Data: []byte{1, 2, 3}},          // 3 bytes
				{ID: "img2", Data: []byte{4, 5, 6, 7, 8, 9}}, // 6 bytes
			},
		},
	}

	summary := buildInspectSummary(doc, nil)
	if summary.TotalMarkdownBytes != 12 {
		t.Errorf("expected TotalMarkdownBytes 12, got %d", summary.TotalMarkdownBytes)
	}
	if summary.TotalMediaBytes != 9 {
		t.Errorf("expected TotalMediaBytes 9, got %d", summary.TotalMediaBytes)
	}
}

func jsonContains(jsonStr, substr string) bool {
	// Simple substring check for compact JSON
	compact := compactJSON(jsonStr)
	return len(compact) > 0 && contains(compact, substr)
}

func compactJSON(s string) string {
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return s
	}
	b, _ := json.Marshal(v)
	return string(b)
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && searchStr(s, sub)
}

func searchStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
