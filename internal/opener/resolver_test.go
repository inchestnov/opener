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
			action, err := Resolve("", []string{tt.target}, cfg)
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

func TestResolve_AliasApp(t *testing.T) {
	cfg := &config.Config{
		Aliases: map[string]config.Rule{
			"ide": {App: "Visual Studio Code"},
		},
	}

	action, err := Resolve("ide", []string{"."}, cfg)
	if err != nil {
		t.Fatalf("Resolve() error = %v, want nil", err)
	}
	if action.Strategy != StrategyApp {
		t.Errorf("Strategy = %v, want StrategyApp", action.Strategy)
	}
	if action.Name != "Visual Studio Code" {
		t.Errorf("Name = %q, want %q", action.Name, "Visual Studio Code")
	}
	if len(action.Args) != 1 || action.Args[0] != "." {
		t.Errorf("Args = %v, want [.]", action.Args)
	}
}

func TestResolve_AliasCommand(t *testing.T) {
	cfg := &config.Config{
		Aliases: map[string]config.Rule{
			"editor": {Command: "nvim"},
		},
	}

	action, err := Resolve("editor", []string{"README.md"}, cfg)
	if err != nil {
		t.Fatalf("Resolve() error = %v, want nil", err)
	}
	if action.Strategy != StrategyCommand {
		t.Errorf("Strategy = %v, want StrategyCommand", action.Strategy)
	}
	if action.Name != "nvim" {
		t.Errorf("Name = %q, want %q", action.Name, "nvim")
	}
	if len(action.Args) != 1 || action.Args[0] != "README.md" {
		t.Errorf("Args = %v, want [README.md]", action.Args)
	}
}

func TestResolve_AliasMultipleTargets(t *testing.T) {
	cfg := &config.Config{
		Aliases: map[string]config.Rule{
			"editor": {Command: "nvim"},
		},
	}

	action, err := Resolve("editor", []string{"a.md", "b.md"}, cfg)
	if err != nil {
		t.Fatalf("Resolve() error = %v, want nil", err)
	}
	if len(action.Args) != 2 || action.Args[0] != "a.md" || action.Args[1] != "b.md" {
		t.Errorf("Args = %v, want [a.md b.md]", action.Args)
	}
}

func TestResolve_UnknownAlias(t *testing.T) {
	cfg := &config.Config{}

	_, err := Resolve("foo", []string{"."}, cfg)
	if err == nil {
		t.Fatal("Resolve() error = nil, want error for unknown alias")
	}
	if got, want := err.Error(), "unknown alias: foo"; got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
}
