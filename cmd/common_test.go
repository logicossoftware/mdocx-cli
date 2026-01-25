package cmd

import "testing"

func TestSanitizeContainerPath_AllowsNormal(t *testing.T) {
	got, err := sanitizeContainerPath("docs/readme.md")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != "docs/readme.md" {
		t.Fatalf("expected %q, got %q", "docs/readme.md", got)
	}

	got, err = sanitizeContainerPath("docs//readme.md")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != "docs/readme.md" {
		t.Fatalf("expected cleaned path %q, got %q", "docs/readme.md", got)
	}
}

func TestSanitizeContainerPath_RejectsTraversalAndAbsolute(t *testing.T) {
	cases := []string{
		"../evil.md",
		"a/../evil.md",
		"a/../../evil.md",
		"/abs.md",
		"C:/abs.md",
		"C:abs.md",
		"\\abs.md",
		"a\\b.md",
		"http://example.com/x",
		"",
		"  ",
	}
	for _, in := range cases {
		_, err := sanitizeContainerPath(in)
		if err == nil {
			t.Fatalf("expected error for %q", in)
		}
	}
}
