package base

import (
	"slices"
	"testing"
)

func TestParse(t *testing.T) {
	t.Setenv("HOME", "/home/tester")
	t.Setenv("SRC_ROOT", "/home/tester/workspace")

	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{"plain path", "/ws", "/ws", false},
		{"trailing separator cleaned", "/ws/", "/ws", false},
		{"env var expanded", "$SRC_ROOT", "/home/tester/workspace", false},
		{"tilde expanded", "~/workspace", "/home/tester/workspace", false},
		{"empty is rejected", "", "", true},
		{"whitespace is rejected", "   ", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := Parse(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse(%q) error = %v, wantErr %v", tt.raw, err, tt.wantErr)
			}
			if b.String() != tt.want {
				t.Errorf("Parse(%q) = %q, want %q", tt.raw, b.String(), tt.want)
			}
		})
	}
}

func TestZeroValueIsPassThrough(t *testing.T) {
	var b Base

	if b.IsSet() {
		t.Error("zero Base IsSet() = true, want false")
	}
	if got := b.Anchor("catalog"); got != "catalog" {
		t.Errorf("Anchor on unset base = %q, want catalog", got)
	}
	if got := b.Shorten("/ws/catalog"); got != "/ws/catalog" {
		t.Errorf("Shorten on unset base = %q, want unchanged", got)
	}
	if got := b.ResolveRoots([]string{"."}); !slices.Equal(got, []string{"."}) {
		t.Errorf("ResolveRoots on unset base = %v, want [.]", got)
	}
}

func TestAnchor(t *testing.T) {
	b := Must("/ws")

	tests := []struct {
		name   string
		target string
		want   string
	}{
		{"bare name is joined", "catalog", "/ws/catalog"},
		{"nested path is joined", "core-apps/api", "/ws/core-apps/api"},
		{"dotfile is an ordinary relative target", ".zshrc", "/ws/.zshrc"},
		{"absolute escapes base", "/etc/hosts", "/etc/hosts"},
		{"tilde escapes base", "~/notes.md", "~/notes.md"},
		{"bare tilde escapes base", "~", "~"},
		{"dot-slash escapes base", "./local", "./local"},
		{"parent escapes base", "../sibling", "../sibling"},
		{"dot escapes base", ".", "."},
		{"parent-dir escapes base", "..", ".."},
		{"URL escapes base", "https://example.com", "https://example.com"},
		{"empty stays empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := b.Anchor(tt.target); got != tt.want {
				t.Errorf("Anchor(%q) = %q, want %q", tt.target, got, tt.want)
			}
		})
	}
}

func TestShorten(t *testing.T) {
	b := Must("/ws")

	tests := []struct {
		name      string
		candidate string
		want      string
	}{
		{"path under base is shortened", "/ws/catalog", "catalog"},
		{"nested path under base is shortened", "/ws/core-apps/api", "core-apps/api"},
		{"base itself is left alone", "/ws", "/ws"},
		{"sibling sharing a prefix is left alone", "/ws-mirror/env", "/ws-mirror/env"},
		{"path outside base keeps its full form", "/elsewhere/other", "/elsewhere/other"},
		{"URL keeps its full form", "https://example.com", "https://example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := b.Shorten(tt.candidate); got != tt.want {
				t.Errorf("Shorten(%q) = %q, want %q", tt.candidate, got, tt.want)
			}
		})
	}
}

func TestResolveRoots(t *testing.T) {
	tests := []struct {
		name  string
		roots []string
		base  Base
		want  []string
	}{
		{"no base leaves roots alone", []string{"."}, Base{}, []string{"."}},
		{"no base, no roots", nil, Base{}, nil},
		{"omitted roots walk the base", nil, Must("/ws"), []string{"/ws"}},
		{"empty roots walk the base", []string{}, Must("/ws"), []string{"/ws"}},
		{"relative root anchors to base", []string{"core-apps"}, Must("/ws"), []string{"/ws/core-apps"}},
		{"dot root is the base", []string{"."}, Must("/ws"), []string{"/ws"}},
		{"absolute root wins", []string{"/other"}, Must("/ws"), []string{"/other"}},
		{"mixed", []string{"sub", "/other"}, Must("/ws"), []string{"/ws/sub", "/other"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.base.ResolveRoots(tt.roots); !slices.Equal(got, tt.want) {
				t.Errorf("ResolveRoots(%v) = %v, want %v", tt.roots, got, tt.want)
			}
		})
	}
}

// The invariant the two halves of base exist to keep: what you tab-complete
// is what opens. A candidate Shorten hands to the shell, Anchor must turn
// back into the same path when you press enter.
func TestRoundTrip(t *testing.T) {
	b := Must("/ws")

	underBase := []string{
		"/ws/catalog",
		"/ws/core-apps/api",
		"/ws/deeply/nested/thing",
	}
	for _, cand := range underBase {
		if got := b.Anchor(b.Shorten(cand)); got != cand {
			t.Errorf("Anchor(Shorten(%q)) = %q, want %q", cand, got, cand)
		}
	}

	// Candidates that keep their full form round-trip too: Shorten leaves
	// them, Anchor's escape hatches leave them.
	fullForm := []string{
		"/elsewhere/other",
		"https://example.com",
	}
	for _, cand := range fullForm {
		if got := b.Anchor(b.Shorten(cand)); got != cand {
			t.Errorf("Anchor(Shorten(%q)) = %q, want %q", cand, got, cand)
		}
	}
}
