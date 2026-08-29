// Package config loads opener's user configuration from ~/.opener.yaml.
package config

import (
	"errors"
	"os"

	"github.com/spf13/viper"
)

// Rule describes how to open something: either a macOS GUI application
// (App) or a CLI executable (Cmd).
//
// On an alias, Complete additionally steers shell completion of that
// alias's targets: "" / "any" (files and directories), "files" (optionally
// narrowed to Extensions), "dirs", "git-dirs" (directories containing a
// .git entry), or "urls" (offer URLs, then fall back to files). Complete,
// Extensions, and URLs are ignored outside alias rules.
type Rule struct {
	App        string   `mapstructure:"app"`
	Cmd        string   `mapstructure:"cmd"`
	Complete   string   `mapstructure:"complete"`
	Extensions []string `mapstructure:"extensions"`
	URLs       []string `mapstructure:"urls"`
}

// PatternRule matches a target's filename against Pattern (a glob, e.g.
// "*.pdf", or a bare extension like ".pdf") and opens it either as a
// macOS GUI application (App) or as a full command line (Cmd, e.g.
// "open -a 'Google Chrome'" - shell-word-split, never run through a shell).
type PatternRule struct {
	Pattern string `mapstructure:"pattern"`
	App     string `mapstructure:"app"`
	Cmd     string `mapstructure:"cmd"`
}

// OpenConfig configures automatic-mode resolution.
type OpenConfig struct {
	Directory Rule          `mapstructure:"directory"`
	Patterns  []PatternRule `mapstructure:"patterns"`
}

// Template is a named shortcut to a fixed file or directory (Path), opened
// either as a macOS GUI application (App) or a CLI executable (Cmd).
// `opener <name>` launches Path, ignoring whatever exists on disk at <name>.
type Template struct {
	Path string `mapstructure:"path"`
	App  string `mapstructure:"app"`
	Cmd  string `mapstructure:"cmd"`
}

// Config is the parsed contents of ~/.opener.yaml.
type Config struct {
	Aliases   map[string]Rule     `mapstructure:"aliases"`
	Templates map[string]Template `mapstructure:"templates"`
	Open      OpenConfig          `mapstructure:"open"`
}

// LoadConfig reads and parses the YAML config file at path. A missing file
// is not an error: it yields a zero-value Config so callers fall back to
// built-in defaults.
func LoadConfig(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
