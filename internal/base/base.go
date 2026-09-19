// Package base models an alias's base directory: the one place its targets
// are rooted.
//
// base is the single setting that spans both halves of opener - shell
// completion offers candidates under it in short, base-relative form, and
// open time joins those short targets back onto it. The join and the trim
// are inverses, and they live here, in one module, rather than mirrored
// across the resolver and source packages where they used to drift apart.
package base

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/inchestnov/opener/internal/pathx"
)

// Base is an alias's base directory. The zero value is the valid "no base"
// state - an alias without a `base:` key - for which every method is a
// pass-through. Build a set Base with Parse.
type Base struct {
	dir string
}

// Parse builds a Base from a raw config value. It expands $VAR / ${VAR} and
// a leading ~, then cleans the result, so everything downstream works with a
// real path.
//
// An empty value is an error: YAML reads a bare `base: ~` as null, and a
// base that silently did nothing would leave completion absolute and targets
// unjoined with nothing to explain why. (A missing `base:` key never reaches
// here - that is the zero Base, not an empty one.)
func Parse(raw string) (Base, error) {
	if strings.TrimSpace(raw) == "" {
		return Base{}, fmt.Errorf("`base:` is empty - YAML reads a bare `~` as null, " +
			`so quote it (base: "~") or use base: $HOME`)
	}
	return Base{dir: filepath.Clean(pathx.Expand(raw))}, nil
}

// Must is Parse for a path already known to be good - a test fixture, or a
// caller holding a real directory. It panics on the empty string.
func Must(raw string) Base {
	b, err := Parse(raw)
	if err != nil {
		panic(err)
	}
	return b
}

// IsSet reports whether b names a directory (the alias had a `base:` key).
func (b Base) IsSet() bool { return b.dir != "" }

// String returns the base directory, or "" when unset. Used in diagnostics.
func (b Base) String() string { return b.dir }

// Anchor joins target onto the base directory, so a short target typed
// against an alias with a base opens the file the completion pointed at.
//
// A target that already says where it lives is returned untouched: an
// absolute path, a ~ path, an explicit ./ or ../ path, a bare . or .., or
// anything carrying a URL scheme. Those escape hatches are what let an alias
// with a base still open a file from somewhere else entirely.
//
// Anchor is the inverse of Shorten: a candidate Shorten made relative,
// Anchor joins back; one Shorten left in full form, Anchor leaves.
func (b Base) Anchor(target string) string {
	if !b.IsSet() {
		return target
	}
	switch {
	case target == "":
		return target
	case filepath.IsAbs(target):
		return target
	case target == "~" || strings.HasPrefix(target, "~/"):
		return target
	case target == "." || target == "..":
		return target
	case strings.HasPrefix(target, "./") || strings.HasPrefix(target, "../"):
		return target
	case strings.Contains(target, "://"):
		return target
	default:
		return filepath.Join(b.dir, target)
	}
}

// Shorten strips the base directory from candidate, so an alias with a base
// completes to the same short paths its targets are written as - with
// absolute candidates, typing a bare name would match nothing.
//
// The match is on a path-segment boundary, not a raw string prefix, so a
// sibling directory sharing a name prefix is left alone. A candidate that
// does not live under base - a URL, or a path from a source spanning several
// roots - keeps its full form and stays completable that way.
func (b Base) Shorten(candidate string) string {
	if !b.IsSet() {
		return candidate
	}
	if rest := strings.TrimPrefix(candidate, b.dir+string(os.PathSeparator)); rest != candidate {
		return rest
	}
	return candidate
}

// ResolveRoots anchors a source's walk roots to the base directory, so
// `base:` means one thing throughout an alias.
//
// Omitting roots entirely walks base itself - an alias that already names
// its directory in `base:` should not have to repeat it. A relative root is
// resolved against base rather than the current directory, which is both the
// useful reading and the safe one: a cwd-relative candidate would be joined
// back onto base at open time and silently point somewhere else. Absolute
// roots are left alone, and without a base nothing changes.
func (b Base) ResolveRoots(roots []string) []string {
	if !b.IsSet() {
		return roots
	}
	if len(roots) == 0 {
		return []string{b.dir}
	}
	out := make([]string, len(roots))
	for i, root := range roots {
		if filepath.IsAbs(root) {
			out[i] = root
			continue
		}
		out[i] = filepath.Join(b.dir, root)
	}
	return out
}
