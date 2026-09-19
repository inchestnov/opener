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
	"sort"
	"strings"

	"github.com/inchestnov/opener/internal/base"
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
// b is the owning alias's base, or the zero base.Base if it has none.
// Candidates found under it are offered relative to it, matching the short
// form those targets are typed in.
func New(spec config.Source, named map[string]config.Source, b base.Base) (Source, error) {
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

	roots := b.ResolveRoots(spec.Roots)

	switch spec.Kind {
	case "list":
		return &listSource{base: b, items: spec.Items}, nil
	case "files":
		return &walkSource{base: b, roots: roots, depth: depthOr(spec.Depth, 2), emit: wantFiles(normExts(spec.Extensions))}, nil
	case "dirs":
		return &walkSource{base: b, roots: roots, depth: depthOr(spec.Depth, 1), emit: wantDirs}, nil
	case "dirs-with":
		if spec.Marker == "" {
			return nil, errors.New("dirs-with source requires a marker")
		}
		return &walkSource{base: b, roots: roots, depth: depthOr(spec.Depth, 1), emit: wantMarker(spec.Marker)}, nil
	case "command":
		if spec.Run == "" {
			return nil, errors.New("command source requires run")
		}
		return &commandSource{base: b, run: spec.Run, cwd: spec.Cwd}, nil
	case "":
		return nil, errors.New("source has no kind")
	default:
		return nil, fmt.Errorf("unknown source kind: %s", spec.Kind)
	}
}

// listSource offers a fixed set of paths and/or URLs.
type listSource struct {
	base  base.Base
	items []string
}

func (l *listSource) Candidates(toComplete string) ([]string, error) {
	return filterSort(l.items, toComplete, l.base), nil
}

// filterSort shortens the candidates against b, keeps those that start with
// toComplete, trims blanks, de-duplicates, sorts, and caps the result at
// maxCandidates.
func filterSort(cands []string, toComplete string, b base.Base) []string {
	prefix := pathx.Expand(toComplete)
	seen := make(map[string]struct{}, len(cands))
	var out []string
	for _, c := range cands {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		c = b.Shorten(c)
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
