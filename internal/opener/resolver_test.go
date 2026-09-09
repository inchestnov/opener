package opener

import (
	"slices"
	"strings"
	"testing"

	"github.com/inchestnov/opener/internal/base"
	"github.com/inchestnov/opener/internal/config"
	"github.com/inchestnov/opener/internal/diagnostic"
)

func TestResolve_AliasApp(t *testing.T) {
	cfg := &config.Config{
		Aliases: map[string]config.Alias{
			"ide": {App: "Visual Studio Code"},
		},
	}

	action, err := Resolve("ide", []string{"."}, cfg, diagnostic.Noop)
	if err != nil {
		t.Fatalf("Resolve() error = %v, want nil", err)
	}
	if action.Strategy != StrategyApp || action.Name != "Visual Studio Code" {
		t.Errorf("got Strategy=%v Name=%q, want StrategyApp %q", action.Strategy, action.Name, "Visual Studio Code")
	}
	if !slices.Equal(action.Args, []string{"."}) {
		t.Errorf("Args = %v, want [.]", action.Args)
	}
}

func TestResolve_AliasCommand(t *testing.T) {
	cfg := &config.Config{
		Aliases: map[string]config.Alias{
			"editor": {Cmd: "nvim"},
		},
	}

	action, err := Resolve("editor", []string{"README.md"}, cfg, diagnostic.Noop)
	if err != nil {
		t.Fatalf("Resolve() error = %v, want nil", err)
	}
	if action.Strategy != StrategyCommand || action.Name != "nvim" {
		t.Errorf("got Strategy=%v Name=%q, want StrategyCommand nvim", action.Strategy, action.Name)
	}
	if !slices.Equal(action.Args, []string{"README.md"}) {
		t.Errorf("Args = %v, want [README.md]", action.Args)
	}
}

func TestResolve_AliasMultipleTargets(t *testing.T) {
	cfg := &config.Config{
		Aliases: map[string]config.Alias{
			"editor": {Cmd: "nvim"},
		},
	}

	action, err := Resolve("editor", []string{"a.md", "b.md"}, cfg, diagnostic.Noop)
	if err != nil {
		t.Fatalf("Resolve() error = %v, want nil", err)
	}
	if !slices.Equal(action.Args, []string{"a.md", "b.md"}) {
		t.Errorf("Args = %v, want [a.md b.md]", action.Args)
	}
}

func TestResolve_AliasCmdSplitting(t *testing.T) {
	cfg := &config.Config{
		Aliases: map[string]config.Alias{
			"chrome": {Cmd: "open -a 'Google Chrome'"},
		},
	}

	action, err := Resolve("chrome", []string{"page.html"}, cfg, diagnostic.Noop)
	if err != nil {
		t.Fatalf("Resolve() error = %v, want nil", err)
	}
	if action.Strategy != StrategyCommand || action.Name != "open" {
		t.Errorf("got Strategy=%v Name=%q, want StrategyCommand open", action.Strategy, action.Name)
	}
	if want := []string{"-a", "Google Chrome", "page.html"}; !slices.Equal(action.Args, want) {
		t.Errorf("Args = %v, want %v", action.Args, want)
	}
}

func TestResolve_AliasCmdUnterminatedQuote(t *testing.T) {
	cfg := &config.Config{
		Aliases: map[string]config.Alias{
			"bad": {Cmd: "open -a 'Google Chrome"},
		},
	}

	if _, err := Resolve("bad", []string{"x"}, cfg, diagnostic.Noop); err == nil {
		t.Fatal("Resolve() error = nil, want error for unterminated quote in cmd")
	}
}

func TestResolve_AliasWithNoAppOrCmd(t *testing.T) {
	cfg := &config.Config{
		Aliases: map[string]config.Alias{
			"empty": {},
		},
	}

	if _, err := Resolve("empty", []string{"x"}, cfg, diagnostic.Noop); err == nil {
		t.Fatal("Resolve() error = nil, want error for an alias with neither app nor cmd")
	}
}

func TestResolve_AliasSourceIgnoredAtOpenTime(t *testing.T) {
	cfg := &config.Config{
		Aliases: map[string]config.Alias{
			"code": {
				App:    "Visual Studio Code",
				Source: config.Source{Kind: "dirs", Roots: []string{"/nonexistent"}},
			},
		},
	}

	action, err := Resolve("code", []string{"whatever-typed"}, cfg, diagnostic.Noop)
	if err != nil {
		t.Fatalf("Resolve() error = %v, want nil", err)
	}
	if !slices.Equal(action.Args, []string{"whatever-typed"}) {
		t.Errorf("Args = %v, want [whatever-typed] (target passed through verbatim)", action.Args)
	}
}

// Resolve routes every target through the alias's base; the full join/trim
// matrix is covered in the base package.
func TestResolve_Base(t *testing.T) {
	cfg := &config.Config{
		Aliases: map[string]config.Alias{
			"workspace": {App: "Visual Studio Code", Base: base.Must("/ws")},
		},
	}

	action, err := Resolve("workspace", []string{"catalog", "/etc/hosts"}, cfg, diagnostic.Noop)
	if err != nil {
		t.Fatalf("Resolve() error = %v, want nil", err)
	}
	if want := []string{"/ws/catalog", "/etc/hosts"}; !slices.Equal(action.Args, want) {
		t.Errorf("Args = %v, want %v", action.Args, want)
	}
}

// With a cmd alias, rebased targets must land after the alias's own fixed
// flags, not in place of them.
func TestResolve_BaseWithCmdFlags(t *testing.T) {
	cfg := &config.Config{
		Aliases: map[string]config.Alias{
			"edit": {Cmd: "nvim -p", Base: base.Must("/ws")},
		},
	}

	action, err := Resolve("edit", []string{"a/x.go", "/tmp/y.go"}, cfg, diagnostic.Noop)
	if err != nil {
		t.Fatalf("Resolve() error = %v, want nil", err)
	}
	if got, want := action.Name, "nvim"; got != want {
		t.Errorf("Name = %q, want %q", got, want)
	}
	if want := []string{"-p", "/ws/a/x.go", "/tmp/y.go"}; !slices.Equal(action.Args, want) {
		t.Errorf("Args = %v, want %v", action.Args, want)
	}
}

func TestResolve_UnknownAlias(t *testing.T) {
	cfg := &config.Config{}

	_, err := Resolve("foo", []string{"."}, cfg, diagnostic.Noop)
	if err == nil {
		t.Fatal("Resolve() error = nil, want error for unknown alias")
	}
	if got, want := err.Error(), "unknown alias: foo"; !strings.Contains(got, want) {
		t.Errorf("error = %q, want it to contain %q", got, want)
	}
}
