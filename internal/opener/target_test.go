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

	tests := []struct {
		name   string
		target string
		want   TargetType
	}{
		{"existing file", file, TargetFile},
		{"existing directory", dir, TargetDirectory},
		{"nonexistent path", filepath.Join(dir, "missing"), TargetUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveTargetType(tt.target); got != tt.want {
				t.Errorf("resolveTargetType(%q) = %v, want %v", tt.target, got, tt.want)
			}
		})
	}
}
