package opener

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/inchestnov/opener/internal/config"
)

func TestResolve_FallbackForUnruledTargets(t *testing.T) {
	dir := t.TempDir()

	file := filepath.Join(dir, "image.png")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatalf("failed to write test fixture: %v", err)
	}
	subdir := filepath.Join(dir, "project")
	if err := os.Mkdir(subdir, 0o755); err != nil {
		t.Fatalf("failed to create test fixture dir: %v", err)
	}

	cfg := &config.Config{}

	tests := []struct {
		name   string
		target string
	}{
		{"existing file with no rule", file},
		{"existing directory", subdir},
		{"nonexistent path / URL", "https://github.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, err := Resolve(tt.target, cfg)
			if err != nil {
				t.Fatalf("Resolve(%q) error = %v, want nil", tt.target, err)
			}
			if action.Strategy != StrategyFallback {
				t.Errorf("Strategy = %v, want StrategyFallback", action.Strategy)
			}
			if len(action.Args) != 1 || action.Args[0] != tt.target {
				t.Errorf("Args = %v, want [%q]", action.Args, tt.target)
			}
		})
	}
}
