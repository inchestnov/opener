package source

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// emitFunc decides, for one entry visited during a walk, whether to emit it
// as a candidate and whether to stop descending into it (directories only).
type emitFunc func(path string, d fs.DirEntry) (take, skip bool)

// walkSource lists filesystem candidates under one or more roots. It backs
// the "files", "dirs", and "dirs-with" kinds via different emit funcs.
//
// A root that is absolute or starts with ~ yields absolute candidates; a
// relative root yields candidates relative to the current directory. depth
// limits how many levels below a root are visited (1 = direct children); a
// negative depth is unlimited. Hidden directories (names starting with ".")
// are never descended into.
type walkSource struct {
	roots []string
	depth int
	emit  emitFunc
}

func (w *walkSource) Candidates(toComplete string) ([]string, error) {
	roots := w.roots
	if len(roots) == 0 {
		roots = []string{"."}
	}

	var out []string
	for _, root := range roots {
		fsRoot := expandUser(root)
		if fsRoot == "" {
			fsRoot = "."
		}

		_ = filepath.WalkDir(fsRoot, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				if path == fsRoot {
					return err // an unreadable root: skip it, try the next
				}
				return nil
			}
			if path == fsRoot {
				return nil // never emit the root itself
			}

			rel, err := filepath.Rel(fsRoot, path)
			if err != nil {
				return nil
			}
			level := strings.Count(rel, string(os.PathSeparator)) + 1

			if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}

			take, skip := w.emit(path, d)
			if take {
				out = append(out, path)
			}
			if d.IsDir() && (skip || (w.depth >= 0 && level >= w.depth)) {
				return filepath.SkipDir
			}
			return nil
		})
	}
	return filterSort(out, toComplete), nil
}

// wantFiles emits regular files, narrowed to the given extensions when the
// list is non-empty.
func wantFiles(exts []string) emitFunc {
	return func(path string, d fs.DirEntry) (bool, bool) {
		if d.IsDir() {
			return false, false
		}
		if len(exts) == 0 {
			return true, false
		}
		ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
		return slices.Contains(exts, ext), false
	}
}

// wantDirs emits every directory.
func wantDirs(_ string, d fs.DirEntry) (bool, bool) {
	return d.IsDir(), false
}

// wantMarker emits a directory that directly contains an entry named
// marker (e.g. ".git"), and does not descend into a match.
func wantMarker(marker string) emitFunc {
	return func(path string, d fs.DirEntry) (bool, bool) {
		if !d.IsDir() {
			return false, false
		}
		if _, err := os.Stat(filepath.Join(path, marker)); err != nil {
			return false, false
		}
		return true, true
	}
}
