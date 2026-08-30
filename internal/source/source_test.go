package source

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/inchestnov/opener/internal/config"
)

func TestNew(t *testing.T) {
	named := map[string]config.Source{
		"real":    {Kind: "list", Items: []string{"x"}},
		"chained": {Ref: "real"},
	}

	tests := []struct {
		name    string
		spec    config.Source
		wantErr bool
	}{
		{"named ref resolves", config.Source{Ref: "real"}, false},
		{"unknown ref", config.Source{Ref: "missing"}, true},
		{"chained ref rejected", config.Source{Ref: "chained"}, true},
		{"inline list", config.Source{Kind: "list"}, false},
		{"unknown kind", config.Source{Kind: "wat"}, true},
		{"empty kind", config.Source{}, true},
		{"dirs-with needs marker", config.Source{Kind: "dirs-with"}, true},
		{"command needs run", config.Source{Kind: "command"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.spec, named)
			if (err != nil) != tt.wantErr {
				t.Errorf("New(%+v) error = %v, wantErr %v", tt.spec, err, tt.wantErr)
			}
		})
	}
}

func TestListSource(t *testing.T) {
	t.Setenv("HOME", "/home/tester")
	s := &listSource{items: []string{"https://z.example", "https://a.example", "~/notes"}}

	got, err := s.Candidates("")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"/home/tester/notes", "https://a.example", "https://z.example"}
	if !slices.Equal(got, want) {
		t.Errorf("Candidates(\"\") = %v, want %v", got, want)
	}

	got, _ = s.Candidates("https://z")
	if !slices.Equal(got, []string{"https://z.example"}) {
		t.Errorf("Candidates(prefix) = %v, want [https://z.example]", got)
	}
}

func TestWalkSource_Files(t *testing.T) {
	root := t.TempDir()
	mkfiles(t, root,
		"a.go", "b.md", "c.txt",
		"sub/d.go", "sub/e.md",
		"sub/deep/f.go",
		".hidden/g.go",
	)

	s := &walkSource{roots: []string{root}, depth: 2, emit: wantFiles([]string{"go", "md"})}
	got, err := s.Candidates("")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		filepath.Join(root, "a.go"),
		filepath.Join(root, "b.md"),
		filepath.Join(root, "sub/d.go"),
		filepath.Join(root, "sub/e.md"),
	}
	if !slices.Equal(got, want) {
		t.Errorf("files = %v\nwant %v", got, want)
	}
}

func TestWalkSource_FilesNoExtensionFilter(t *testing.T) {
	root := t.TempDir()
	mkfiles(t, root, "a.go", "README")

	s := &walkSource{roots: []string{root}, depth: 1, emit: wantFiles(nil)}
	got, _ := s.Candidates("")
	want := []string{filepath.Join(root, "README"), filepath.Join(root, "a.go")}
	if !slices.Equal(got, want) {
		t.Errorf("files = %v, want %v", got, want)
	}
}

func TestWalkSource_Dirs(t *testing.T) {
	root := t.TempDir()
	mkfiles(t, root, "one/x", "two/y", "two/nested/z")

	s := &walkSource{roots: []string{root}, depth: 1, emit: wantDirs}
	got, _ := s.Candidates("")
	want := []string{filepath.Join(root, "one"), filepath.Join(root, "two")}
	if !slices.Equal(got, want) {
		t.Errorf("dirs = %v, want %v", got, want)
	}
}

func TestWalkSource_DirsWith(t *testing.T) {
	root := t.TempDir()
	for _, d := range []string{"repo-a/.git", "repo-b/.git", "plain", "nested/repo-c/.git", "repo-a/vendor/dep/.git"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	s := &walkSource{roots: []string{root}, depth: 1, emit: wantMarker(".git")}
	got, _ := s.Candidates("")
	want := []string{filepath.Join(root, "repo-a"), filepath.Join(root, "repo-b")}
	if !slices.Equal(got, want) {
		t.Errorf("dirs-with (depth 1) = %v, want %v", got, want)
	}

	s = &walkSource{roots: []string{root}, depth: 2, emit: wantMarker(".git")}
	got, _ = s.Candidates("")
	want = []string{filepath.Join(root, "nested/repo-c"), filepath.Join(root, "repo-a"), filepath.Join(root, "repo-b")}
	if !slices.Equal(got, want) {
		t.Errorf("dirs-with (depth 2) = %v, want %v", got, want)
	}
}

func TestWalkSource_RelativeRootStaysRelative(t *testing.T) {
	root := t.TempDir()
	mkfiles(t, root, "sub/a.go")
	chdir(t, root)

	s := &walkSource{roots: []string{"."}, depth: 2, emit: wantFiles([]string{"go"})}
	got, _ := s.Candidates("")
	if !slices.Equal(got, []string{"sub/a.go"}) {
		t.Errorf("relative-root candidates = %v, want [sub/a.go]", got)
	}
}

func TestWalkSource_PrefixFilter(t *testing.T) {
	root := t.TempDir()
	mkfiles(t, root, "alpha.go", "beta.go")

	s := &walkSource{roots: []string{root}, depth: 1, emit: wantFiles([]string{"go"})}
	got, _ := s.Candidates(filepath.Join(root, "al"))
	if !slices.Equal(got, []string{filepath.Join(root, "alpha.go")}) {
		t.Errorf("prefix-filtered = %v, want [alpha.go]", got)
	}
}

func TestCommandSource(t *testing.T) {
	s := &commandSource{run: `printf 'c\na\nb\n'`}
	got, err := s.Candidates("")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, []string{"a", "b", "c"}) {
		t.Errorf("command candidates = %v, want [a b c]", got)
	}

	got, _ = s.Candidates("b")
	if !slices.Equal(got, []string{"b"}) {
		t.Errorf("prefix-filtered = %v, want [b]", got)
	}

	if _, err := (&commandSource{run: "exit 1"}).Candidates(""); err == nil {
		t.Error("Candidates() error = nil, want error for failing command")
	}
}

func TestCommandSource_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("timeout test sleeps for the command timeout")
	}
	_, err := (&commandSource{run: "sleep 5"}).Candidates("")
	if err == nil {
		t.Error("Candidates() error = nil, want timeout error")
	}
}

func TestExpandUser(t *testing.T) {
	t.Setenv("HOME", "/home/tester")
	cases := map[string]string{
		"~":             "/home/tester",
		"~/code":        "/home/tester/code",
		"relative/path": "relative/path",
		"/absolute":     "/absolute",
		"~user/thing":   "~user/thing",
	}
	for in, want := range cases {
		if got := expandUser(in); got != want {
			t.Errorf("expandUser(%q) = %q, want %q", in, got, want)
		}
	}
}

func mkfiles(t *testing.T, root string, rels ...string) {
	t.Helper()
	for _, rel := range rels {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(prev) })
}
