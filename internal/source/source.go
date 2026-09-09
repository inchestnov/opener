// Package source discovers shell-completion candidates for an alias's
// targets. A Source is built from the user's config (an inline spec or a
// reference into the top-level `sources:` map) and is consulted only during
// completion, never when a target is actually opened.
//
// Paths arriving from the config have already been expanded by
// config.LoadConfig, so this package works with real paths throughout. The
// one exception is the partial target being completed, which comes from the
// user's keyboard and is expanded here.
package source

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/inchestnov/opener/internal/config"
	"github.com/inchestnov/opener/internal/pathx"
)

// maxCandidates caps how many completions a source returns, so a broad
// `command` or file walk can't flood the shell.
const maxCandidates = 2000

// Source produces completion candidates for the target currently being
// typed. Each candidate is a full path or full URL.
type Source interface {
	Candidates(toComplete string) ([]string, error)
}

// New builds a Source from spec. A spec with Ref set is resolved against
// named (the config's top-level `sources:` map); an inline spec switches on
// Kind. Referencing another reference is rejected.
//
// base is the owning alias's `base:`, or "" if it has none. Candidates
// found under base are offered relative to it, matching the short form
// those targets are typed in.
func New(spec config.Source, named map[string]config.Source, base string) (Source, error) {
	if spec.Ref != "" {
		s, ok := named[spec.Ref]
		if !ok {
			return nil, fmt.Errorf("unknown source: %s", spec.Ref)
		}
		if s.Ref != "" {
			return nil, fmt.Errorf("source %q refers to another source; chaining is not allowed", spec.Ref)
		}
		spec = s
	}

	roots := resolveRoots(spec.Roots, base)

	switch spec.Kind {
	case "list":
		return &listSource{base: base, items: spec.Items}, nil
	case "files":
		return &walkSource{base: base, roots: roots, depth: depthOr(spec.Depth, 2), emit: wantFiles(normExts(spec.Extensions))}, nil
	case "dirs":
		return &walkSource{base: base, roots: roots, depth: depthOr(spec.Depth, 1), emit: wantDirs}, nil
	case "dirs-with":
		if spec.Marker == "" {
			return nil, errors.New("dirs-with source requires a marker")
		}
		return &walkSource{base: base, roots: roots, depth: depthOr(spec.Depth, 1), emit: wantMarker(spec.Marker)}, nil
	case "command":
		if spec.Run == "" {
			return nil, errors.New("command source requires run")
		}
		return &commandSource{base: base, run: spec.Run, cwd: spec.Cwd}, nil
	case "":
		return nil, errors.New("source has no kind")
	default:
		return nil, fmt.Errorf("unknown source kind: %s", spec.Kind)
	}
}

// listSource offers a fixed set of paths and/or URLs.
type listSource struct {
	base  string
	items []string
}

func (l *listSource) Candidates(toComplete string) ([]string, error) {
	return filterSort(l.items, toComplete, l.base), nil
}

// filterSort rebases the candidates onto base, keeps those that start with
// toComplete, trims blanks, de-duplicates, sorts, and caps the result at
// maxCandidates.
func filterSort(cands []string, toComplete, base string) []string {
	prefix := pathx.Expand(toComplete)
	seen := make(map[string]struct{}, len(cands))
	var out []string
	for _, c := range cands {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		c = rebase(c, base)
		if !strings.HasPrefix(c, prefix) {
			continue
		}
		if _, dup := seen[c]; dup {
			continue
		}
		seen[c] = struct{}{}
		out = append(out, c)
	}
	sort.Strings(out)
	if len(out) > maxCandidates {
		out = out[:maxCandidates]
	}
	return out
}

// resolveRoots anchors a walk's roots to base, so `base:` means one thing
// throughout an alias: the directory its paths are relative to.
//
// Omitting roots entirely walks base itself - an alias that already names
// its directory in `base:` should not have to repeat it. A relative root is
// resolved against base rather than the current directory, which is both
// the useful reading and the safe one: a cwd-relative candidate would be
// joined back onto base at open time and silently point somewhere else.
// Absolute roots are left alone, and without a base nothing changes.
func resolveRoots(roots []string, base string) []string {
	if base == "" {
		return roots
	}
	if len(roots) == 0 {
		return []string{base}
	}
	out := make([]string, len(roots))
	for i, root := range roots {
		if filepath.IsAbs(root) {
			out[i] = root
			continue
		}
		out[i] = filepath.Join(base, root)
	}
	return out
}

// rebase strips base from candidate c, so an alias with a base completes to
// the same short paths its targets are written as. A candidate that does
// not live under base - a URL, or a path from a source spanning several
// roots - keeps its full form, and stays completable by typing that form.
func rebase(c, base string) string {
	if base == "" {
		return c
	}
	if rest := strings.TrimPrefix(c, base+string(os.PathSeparator)); rest != c {
		return rest
	}
	return c
}

// normExts lowercases each extension and strips a leading dot, dropping
// empties (".GO" -> "go").
func normExts(exts []string) []string {
	out := make([]string, 0, len(exts))
	for _, e := range exts {
		if e = strings.TrimPrefix(strings.TrimSpace(strings.ToLower(e)), "."); e != "" {
			out = append(out, e)
		}
	}
	return out
}

// depthOr returns *d, or def when d is nil (the key was omitted).
func depthOr(d *int, def int) int {
	if d == nil {
		return def
	}
	return *d
}
