package config

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestLoadConfig_MissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.yaml")

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}
	if cfg == nil {
		t.Fatal("LoadConfig() returned nil Config, want empty Config")
	}
	if len(cfg.Aliases) != 0 {
		t.Errorf("Aliases = %v, want empty", cfg.Aliases)
	}
	if len(cfg.Sources) != 0 {
		t.Errorf("Sources = %v, want empty", cfg.Sources)
	}
}

func TestLoadConfig_ValidFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "opener.yaml")
	writeFile(t, path, `
sources:
  repos:
    kind: dirs-with
    roots: ["~/src", "~/work"]
    marker: .git
    depth: 2
  notes:
    kind: files
    roots: ["~/notes"]
    extensions: [md, txt]
  bookmarks:
    kind: list
    items:
      - https://github.com
  gh:
    kind: command
    run: "gh repo list"

aliases:
  code:
    app: "Visual Studio Code"
    source: repos
  edit:
    cmd: "nvim -p"
    source:
      kind: files
      extensions: [go]
  open:
    cmd: open
`)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}

	if got, want := cfg.Aliases["code"].App, "Visual Studio Code"; got != want {
		t.Errorf("Aliases[code].App = %q, want %q", got, want)
	}
	if got, want := cfg.Aliases["code"].Source.Ref, "repos"; got != want {
		t.Errorf("Aliases[code].Source.Ref = %q, want %q", got, want)
	}
	if got, want := cfg.Aliases["edit"].Cmd, "nvim -p"; got != want {
		t.Errorf("Aliases[edit].Cmd = %q, want %q", got, want)
	}
	if got, want := cfg.Aliases["edit"].Source.Kind, "files"; got != want {
		t.Errorf("Aliases[edit].Source.Kind = %q, want %q", got, want)
	}
	if got, want := cfg.Aliases["edit"].Source.Extensions, []string{"go"}; !slices.Equal(got, want) {
		t.Errorf("Aliases[edit].Source.Extensions = %v, want %v", got, want)
	}
	if !cfg.Aliases["open"].Source.IsZero() {
		t.Errorf("Aliases[open].Source = %+v, want zero", cfg.Aliases["open"].Source)
	}

	repos := cfg.Sources["repos"]
	if repos.Kind != "dirs-with" || repos.Marker != ".git" {
		t.Errorf("Sources[repos] = %+v, want kind=dirs-with marker=.git", repos)
	}
	if repos.Depth == nil || *repos.Depth != 2 {
		t.Errorf("Sources[repos].Depth = %v, want 2", repos.Depth)
	}
	if got, want := cfg.Sources["repos"].Roots, []string{"~/src", "~/work"}; !slices.Equal(got, want) {
		t.Errorf("Sources[repos].Roots = %v, want %v", got, want)
	}
	if got, want := cfg.Sources["gh"].Run, "gh repo list"; got != want {
		t.Errorf("Sources[gh].Run = %q, want %q", got, want)
	}
}

func TestLoadConfig_SourceStringVsMapping(t *testing.T) {
	path := filepath.Join(t.TempDir(), "opener.yaml")
	writeFile(t, path, `
aliases:
  named:
    app: X
    source: my-source
  inline:
    app: Y
    source:
      kind: dirs
      roots: ["."]
`)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}

	if got := cfg.Aliases["named"].Source; got.Ref != "my-source" || got.Kind != "" {
		t.Errorf("named source = %+v, want Ref=my-source only", got)
	}
	if got := cfg.Aliases["inline"].Source; got.Ref != "" || got.Kind != "dirs" {
		t.Errorf("inline source = %+v, want Kind=dirs, no Ref", got)
	}
}

func TestLoadConfig_RejectsRemovedKeys(t *testing.T) {
	tests := map[string]string{
		"templates": "templates:\n  x:\n    path: /a\n",
		"open":      "open:\n  directory:\n    app: Finder\n",
	}
	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "opener.yaml")
			writeFile(t, path, body)

			_, err := LoadConfig(path)
			if err == nil {
				t.Fatalf("LoadConfig() error = nil, want error for `%s:`", name)
			}
			if !strings.Contains(err.Error(), name) {
				t.Errorf("error = %q, want it to mention %q", err, name)
			}
		})
	}
}

func TestLoadConfig_RejectsUnknownAliasKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "opener.yaml")
	writeFile(t, path, "aliases:\n  x:\n    app: X\n    complete: dirs\n")

	if _, err := LoadConfig(path); err == nil {
		t.Fatal("LoadConfig() error = nil, want error for stale `complete:` key")
	}
}

func TestLoadConfig_MalformedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "opener.yaml")
	writeFile(t, path, "aliases: [this is not: a valid map")

	if _, err := LoadConfig(path); err == nil {
		t.Fatal("LoadConfig() error = nil, want error for malformed YAML")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test fixture: %v", err)
	}
}
