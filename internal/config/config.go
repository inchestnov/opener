// Package config loads opener's user configuration from ~/.opener.yaml.
package config

import (
	"errors"
	"os"

	"github.com/spf13/viper"
)

// Rule describes how to open something: either a macOS GUI application
// (App) or a CLI executable (Cmd).
type Rule struct {
	App string `mapstructure:"app"`
	Cmd string `mapstructure:"cmd"`
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

// Config is the parsed contents of ~/.opener.yaml.
type Config struct {
	Aliases map[string]Rule `mapstructure:"aliases"`
	Open    OpenConfig      `mapstructure:"open"`
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
