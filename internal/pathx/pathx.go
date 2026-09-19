// Package pathx expands the shell-style shorthands opener accepts in
// configured paths: environment variables and a leading ~.
//
// Config values are expanded once, when the config is loaded, so everything
// downstream works with real paths.
package pathx

import (
	"os"
	"strings"
)

// Expand resolves $VAR and ${VAR} references in p against the environment,
// then a leading ~ or ~/ against the home directory. Expansion happens in
// that order, so a variable holding a ~ path expands the whole way.
//
// An unset variable is left in place verbatim rather than expanding to the
// empty string: a typo like $SRC_ROT then surfaces as a path that visibly
// does not exist, instead of silently collapsing "$SRC_ROT/core-apps" into
// "/core-apps".
func Expand(p string) string {
	if strings.ContainsRune(p, '$') {
		p = os.Expand(p, func(name string) string {
			if v, ok := os.LookupEnv(name); ok {
				return v
			}
			return "$" + name
		})
	}
	return expandUser(p)
}

// expandUser resolves a leading ~ or ~/ against the home directory.
func expandUser(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return home + strings.TrimPrefix(p, "~")
		}
	}
	return p
}
