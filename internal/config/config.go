// Package config loads opener's user configuration from ~/.opener.yaml.
package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"

	"github.com/inchestnov/opener/internal/base"
	"github.com/inchestnov/opener/internal/pathx"
)

// Alias is a named launcher for `opener <alias> <target>...`: it opens its
// targets as a macOS GUI application (App) or as a CLI command line (Cmd).
//
// Cmd is split into words the way a shell would (single/double quotes
// honored) and run directly - no shell is ever invoked - with the targets
// appended, so `cmd: "open -a 'Google Chrome'"` works.
//
// Source, when set, drives shell completion of this alias's targets.
//
// Base, when set, is the directory this alias's targets are written
// relative to. It is the one setting that affects both halves: completion
// candidates under Base are offered in their short, Base-relative form, and
// at open time a target that is not anchored elsewhere is joined back onto
// it. Without Base, targets are passed through exactly as typed. See the
// base package for the join/trim rules.
type Alias struct {
	App    string    `mapstructure:"app"`
	Cmd    string    `mapstructure:"cmd"`
	Base   base.Base `mapstructure:"base"`
	Source Source    `mapstructure:"source"`
}

// Source is a target-discovery spec used for shell completion. In YAML it is
// either a bare string naming an entry in the top-level `sources:` map
// (decoded into Ref) or an inline mapping (Kind plus the fields that kind
// uses).
//
// Kinds:
//   - "list": Items, a fixed set of paths and/or URLs.
//   - "files": files under Roots (default ["."]), narrowed to Extensions,
//     no deeper than Depth levels (default 2).
//   - "dirs": directories under Roots, no deeper than Depth (default 1).
//   - "dirs-with": directories under Roots that directly contain Marker
//     (e.g. ".git"), no deeper than Depth (default 1).
//   - "command": each line of stdout from running Run via `sh -c` (in Cwd,
//     if set).
type Source struct {
	Ref string `mapstructure:"ref"`

	Kind       string   `mapstructure:"kind"`
	Items      []string `mapstructure:"items"`
	Roots      []string `mapstructure:"roots"`
	Extensions []string `mapstructure:"extensions"`
	Depth      *int     `mapstructure:"depth"`
	Marker     string   `mapstructure:"marker"`
	Run        string   `mapstructure:"run"`
	Cwd        string   `mapstructure:"cwd"`
}

// IsZero reports whether s carries no source configuration at all (the
// alias had no `source:` key). Source is not comparable, so callers use
// this instead of an equality check against the zero value.
func (s Source) IsZero() bool {
	return s.Ref == "" && s.Kind == "" && s.Items == nil && s.Roots == nil &&
		s.Extensions == nil && s.Depth == nil && s.Marker == "" &&
		s.Run == "" && s.Cwd == ""
}

// Config is the parsed contents of ~/.opener.yaml.
type Config struct {
	Aliases map[string]Alias  `mapstructure:"aliases"`
	Sources map[string]Source `mapstructure:"sources"`
}

// removedKeys are top-level keys from earlier opener versions. Seeing one is
// a hard error with a pointer to the current docs, rather than a silently
// ignored setting.
var removedKeys = []string{"templates", "open"}

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

	for _, k := range removedKeys {
		if v.IsSet(k) {
			return nil, fmt.Errorf("`%s:` is no longer supported - opener now uses `aliases:` + `sources:`; "+
				"see https://github.com/inchestnov/opener#configuration", k)
		}
	}

	var cfg Config
	err := v.Unmarshal(&cfg,
		func(dc *mapstructure.DecoderConfig) { dc.ErrorUnused = true },
		viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
			sourceStringHook(),
			baseStringHook(),
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		)),
	)
	if err != nil {
		return nil, err
	}
	if err := rejectNullBase(&cfg, v.Get("aliases")); err != nil {
		return nil, err
	}
	expandPaths(&cfg)
	return &cfg, nil
}

// rejectNullBase rejects an alias that wrote a `base:` key which carried no
// value. The trap is YAML's, not ours: a bare `base: ~` is null and never
// reaches the decode hook, so the setting would silently do nothing -
// completion would stay absolute and targets unjoined, with no error to
// explain why. A `base:` with a real string is already a set base.Base by
// this point (or, if empty, failed in base.Parse during Unmarshal); this
// only has to catch the key-present-but-absent case.
func rejectNullBase(cfg *Config, rawAliases any) error {
	byName, ok := rawAliases.(map[string]any)
	if !ok {
		return nil
	}
	for name, a := range cfg.Aliases {
		if a.Base.IsSet() {
			continue
		}
		fields, ok := byName[name].(map[string]any)
		if !ok {
			continue
		}
		if _, present := fields["base"]; present {
			return fmt.Errorf("alias %q: `base:` is empty - YAML reads a bare `~` as null, "+
				`so quote it (base: "~") or use base: $HOME`, name)
		}
	}
	return nil
}

// expandPaths rewrites every path-valued setting through pathx.Expand, so
// $VAR and ~ are resolved once, here, rather than at each point of use.
// URLs and other non-path values pass through untouched: they contain
// neither a $ nor a leading ~. (An alias's Base is already expanded - it
// went through base.Parse in the decode hook.)
func expandPaths(cfg *Config) {
	for name, a := range cfg.Aliases {
		expandSourcePaths(&a.Source)
		cfg.Aliases[name] = a
	}
	for name, s := range cfg.Sources {
		expandSourcePaths(&s)
		cfg.Sources[name] = s
	}
}

// expandSourcePaths expands the path-valued fields of a single source spec
// in place.
func expandSourcePaths(s *Source) {
	for i, root := range s.Roots {
		s.Roots[i] = pathx.Expand(root)
	}
	for i, item := range s.Items {
		s.Items[i] = pathx.Expand(item)
	}
	s.Cwd = pathx.Expand(s.Cwd)
}

// sourceStringHook lets an alias's `source:` be written as a bare string
// (a reference into the `sources:` map) as well as an inline mapping.
func sourceStringHook() mapstructure.DecodeHookFuncType {
	sourceType := reflect.TypeOf(Source{})
	return func(from, to reflect.Type, data any) (any, error) {
		if to != sourceType || from.Kind() != reflect.String {
			return data, nil
		}
		return Source{Ref: data.(string)}, nil
	}
}

// baseStringHook turns an alias's `base:` string into a base.Base, expanding
// and cleaning it. An empty string fails here; a null (`base: ~`) never
// reaches a hook and is caught by rejectNullBase instead.
func baseStringHook() mapstructure.DecodeHookFuncType {
	baseType := reflect.TypeOf(base.Base{})
	return func(from, to reflect.Type, data any) (any, error) {
		if to != baseType || from.Kind() != reflect.String {
			return data, nil
		}
		return base.Parse(data.(string))
	}
}
