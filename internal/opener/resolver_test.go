package opener

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/inchestnov/opener/internal/config"
	"github.com/inchestnov/opener/internal/diagnostic"
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
			action, err := Resolve("", []string{tt.target}, cfg, diagnostic.Context{Logger: diagnostic.Noop})
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

func TestResolve_DirectoryRuleOverride(t *testing.T) {
	dir := t.TempDir()

	cfg := &config.Config{
		Open: config.OpenConfig{
			Directory: config.Rule{App: "Visual Studio Code"},
		},
	}

	action, err := Resolve("", []string{dir}, cfg, diagnostic.Context{Logger: diagnostic.Noop})
	if err != nil {
		t.Fatalf("Resolve() error = %v, want nil", err)
	}
	if action.Strategy != StrategyApp || action.Name != "Visual Studio Code" {
		t.Errorf("got Strategy=%v Name=%q, want StrategyApp %q", action.Strategy, action.Name, "Visual Studio Code")
	}
}

func TestResolve_AliasApp(t *testing.T) {
	cfg := &config.Config{
		Aliases: map[string]config.Rule{
			"ide": {App: "Visual Studio Code"},
		},
	}

	action, err := Resolve("ide", []string{"."}, cfg, diagnostic.Context{Logger: diagnostic.Noop})
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

	action, err := Resolve("editor", []string{"README.md"}, cfg, diagnostic.Context{Logger: diagnostic.Noop})
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

	action, err := Resolve("editor", []string{"a.md", "b.md"}, cfg, diagnostic.Context{Logger: diagnostic.Noop})
	if err != nil {
		t.Fatalf("Resolve() error = %v, want nil", err)
	}
	if len(action.Args) != 2 || action.Args[0] != "a.md" || action.Args[1] != "b.md" {
		t.Errorf("Args = %v, want [a.md b.md]", action.Args)
	}
}

func TestResolve_UnknownAlias(t *testing.T) {
	cfg := &config.Config{}

	_, err := Resolve("foo", []string{"."}, cfg, diagnostic.Context{Logger: diagnostic.Noop})
	if err == nil {
		t.Fatal("Resolve() error = nil, want error for unknown alias")
	}
	if got, want := err.Error(), "unknown alias: foo"; got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
}

func TestResolve_PatternRuleGlobApp(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "document.pdf")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatalf("failed to write test fixture: %v", err)
	}

	cfg := &config.Config{
		Open: config.OpenConfig{
			Patterns: []config.PatternRule{
				{Pattern: "*.pdf", App: "Google Chrome"},
			},
		},
	}

	action, err := Resolve("", []string{file}, cfg, diagnostic.Context{Logger: diagnostic.Noop})
	if err != nil {
		t.Fatalf("Resolve() error = %v, want nil", err)
	}
	if action.Strategy != StrategyApp {
		t.Errorf("Strategy = %v, want StrategyApp", action.Strategy)
	}
	if action.Name != "Google Chrome" {
		t.Errorf("Name = %q, want %q", action.Name, "Google Chrome")
	}
	if len(action.Args) != 1 || action.Args[0] != file {
		t.Errorf("Args = %v, want [%q]", action.Args, file)
	}
}

func TestResolve_PatternRuleGlobCaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "document.PDF")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatalf("failed to write test fixture: %v", err)
	}

	cfg := &config.Config{
		Open: config.OpenConfig{
			Patterns: []config.PatternRule{
				{Pattern: "*.pdf", App: "Safari"},
			},
		},
	}

	action, err := Resolve("", []string{file}, cfg, diagnostic.Context{Logger: diagnostic.Noop})
	if err != nil {
		t.Fatalf("Resolve() error = %v, want nil", err)
	}
	if action.Strategy != StrategyApp || action.Name != "Safari" {
		t.Errorf("got Strategy=%v Name=%q, want StrategyApp Safari", action.Strategy, action.Name)
	}
}

func TestResolve_PatternRuleBareExtension(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "document.pdf")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatalf("failed to write test fixture: %v", err)
	}

	cfg := &config.Config{
		Open: config.OpenConfig{
			// ".pdf" (no glob metacharacters) behaves like "*.pdf".
			Patterns: []config.PatternRule{
				{Pattern: ".pdf", App: "Safari"},
			},
		},
	}

	action, err := Resolve("", []string{file}, cfg, diagnostic.Context{Logger: diagnostic.Noop})
	if err != nil {
		t.Fatalf("Resolve() error = %v, want nil", err)
	}
	if action.Strategy != StrategyApp || action.Name != "Safari" {
		t.Errorf("got Strategy=%v Name=%q, want StrategyApp Safari", action.Strategy, action.Name)
	}
}

func TestResolve_PatternRuleCmd(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "document.pdf")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatalf("failed to write test fixture: %v", err)
	}

	cfg := &config.Config{
		Open: config.OpenConfig{
			Patterns: []config.PatternRule{
				{Pattern: ".pdf", Cmd: "open -a 'Google Chrome'"},
			},
		},
	}

	action, err := Resolve("", []string{file}, cfg, diagnostic.Context{Logger: diagnostic.Noop})
	if err != nil {
		t.Fatalf("Resolve() error = %v, want nil", err)
	}
	if action.Strategy != StrategyCommand {
		t.Errorf("Strategy = %v, want StrategyCommand", action.Strategy)
	}
	if action.Name != "open" {
		t.Errorf("Name = %q, want %q", action.Name, "open")
	}
	if want := []string{"-a", "Google Chrome", file}; !slicesEqual(action.Args, want) {
		t.Errorf("Args = %v, want %v", action.Args, want)
	}
}

func TestResolve_PatternRuleFirstMatchWins(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "document.pdf")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatalf("failed to write test fixture: %v", err)
	}

	cfg := &config.Config{
		Open: config.OpenConfig{
			Patterns: []config.PatternRule{
				{Pattern: "document.pdf", App: "Safari"},
				{Pattern: "*.pdf", App: "Google Chrome"},
			},
		},
	}

	action, err := Resolve("", []string{file}, cfg, diagnostic.Context{Logger: diagnostic.Noop})
	if err != nil {
		t.Fatalf("Resolve() error = %v, want nil", err)
	}
	if action.Name != "Safari" {
		t.Errorf("Name = %q, want %q (first matching pattern should win)", action.Name, "Safari")
	}
}

func TestResolve_PatternRuleWithoutMatchFallsBack(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "image.png")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatalf("failed to write test fixture: %v", err)
	}

	cfg := &config.Config{
		Open: config.OpenConfig{
			Patterns: []config.PatternRule{
				{Pattern: "*.pdf", App: "Google Chrome"},
			},
		},
	}

	action, err := Resolve("", []string{file}, cfg, diagnostic.Context{Logger: diagnostic.Noop})
	if err != nil {
		t.Fatalf("Resolve() error = %v, want nil", err)
	}
	if action.Strategy != StrategyFallback {
		t.Errorf("Strategy = %v, want StrategyFallback", action.Strategy)
	}
	if len(action.Args) != 1 || action.Args[0] != file {
		t.Errorf("Args = %v, want [%q]", action.Args, file)
	}
}

// Pattern matching isn't hardcoded to pdf: any glob you configure is
// checked, in order, against the target's filename.
func TestResolve_NonPDFPatternRule(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "notes.md")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatalf("failed to write test fixture: %v", err)
	}

	cfg := &config.Config{
		Open: config.OpenConfig{
			Patterns: []config.PatternRule{
				{Pattern: "*.pdf", App: "Google Chrome"},
				{Pattern: "*.md", Cmd: "nvim"},
			},
		},
	}

	action, err := Resolve("", []string{file}, cfg, diagnostic.Context{Logger: diagnostic.Noop})
	if err != nil {
		t.Fatalf("Resolve() error = %v, want nil", err)
	}
	if action.Strategy != StrategyCommand || action.Name != "nvim" {
		t.Errorf("got Strategy=%v Name=%q, want StrategyCommand nvim", action.Strategy, action.Name)
	}
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
