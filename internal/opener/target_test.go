package opener

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveTargetType(t *testing.T) {
	dir := t.TempDir()

	file := filepath.Join(dir, "document.pdf")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatalf("failed to write test fixture: %v", err)
	}

	script := filepath.Join(dir, "run.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh"), 0o700); err != nil {
		t.Fatalf("failed to write test fixture: %v", err)
	}

	tests := []struct {
		name   string
		target string
		want   TargetType
	}{
		{"existing file", file, TargetFile},
		{"existing directory", dir, TargetDirectory},
		{"executable file", script, TargetExecutable},
		{"nonexistent path", filepath.Join(dir, "missing"), TargetUnknown},
		{"url", "https://github.com", TargetURL},
		{"url with non-slash scheme", "mailto:foo@example.com", TargetURL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveTargetType(tt.target); got != tt.want {
				t.Errorf("resolveTargetType(%q) = %v, want %v", tt.target, got, tt.want)
			}
		})
	}
}
