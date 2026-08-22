package config

import (
	"os"
	"path/filepath"
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
	if cfg.Open.Directory.App != "" || cfg.Open.Directory.Command != "" {
		t.Errorf("Open.Directory = %+v, want zero value", cfg.Open.Directory)
	}
	if len(cfg.Open.Files) != 0 {
		t.Errorf("Open.Files = %v, want empty", cfg.Open.Files)
	}
}

func TestLoadConfig_ValidFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "opener.yaml")
	writeFile(t, path, `
aliases:
  ide:
    app: "Visual Studio Code"
  editor:
    command: "nvim"

open:
  directory:
    command: "open"
  files:
    pdf:
      app: "Google Chrome"
`)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}

	if got, want := cfg.Aliases["ide"].App, "Visual Studio Code"; got != want {
		t.Errorf("Aliases[ide].App = %q, want %q", got, want)
	}
	if got, want := cfg.Aliases["editor"].Command, "nvim"; got != want {
		t.Errorf("Aliases[editor].Command = %q, want %q", got, want)
	}
	if got, want := cfg.Open.Directory.Command, "open"; got != want {
		t.Errorf("Open.Directory.Command = %q, want %q", got, want)
	}
	if got, want := cfg.Open.Files["pdf"].App, "Google Chrome"; got != want {
		t.Errorf("Open.Files[pdf].App = %q, want %q", got, want)
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
