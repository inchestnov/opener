package cli

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/spf13/cobra"

	"github.com/inchestnov/opener/internal/config"
)

func TestCompleteArg(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	const cfg = `
aliases:
  code:
    app: "Visual Studio Code"
    complete: git-dirs
  web:
    app: "Google Chrome"
    complete: urls
    urls:
      - https://github.com
      - https://example.com
  edit:
    cmd: nvim
    complete: files
    extensions: [".go", "md"]
  show:
    app: "Finder"
    complete: dirs
  raw:
    cmd: cat
templates:
  zshrc:
    path: ~/.zshrc
  opener-config:
    path: ~/.opener.yaml
`
	if err := os.WriteFile(filepath.Join(home, ".opener.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		args       []string
		toComplete string
		want       []string
	}{
		{
			name:       "first arg, empty prefix: defer to file completion",
			toComplete: "",
			want:       nil,
		},
		{
			name:       "first arg, prefix filters",
			toComplete: "o",
			want:       []string{"opener-config\ttemplate"},
		},
		{
			name:       "first arg, alias prefix",
			toComplete: "c",
			want:       []string{"code\talias"},
		},
		{
			name:       "unknown alias: plain file completion",
			args:       []string{"nope"},
			toComplete: "",
			want:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, directive := completeArg(nil, tt.args, tt.toComplete)
			if directive != cobra.ShellCompDirectiveDefault {
				t.Errorf("directive = %v, want ShellCompDirectiveDefault", directive)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("completeArg(%q, %q) = %q, want %q", tt.args, tt.toComplete, got, tt.want)
			}
		})
	}
}

func TestCompleteArgAliasTarget(t *testing.T) {
	tests := []struct {
		name          string
		rule          config.Rule
		toComplete    string
		want          []string
		wantDirective cobra.ShellCompDirective
	}{
		{
			name:          "dirs",
			rule:          config.Rule{Complete: "dirs"},
			wantDirective: cobra.ShellCompDirectiveFilterDirs,
		},
		{
			name:          "files without extensions",
			rule:          config.Rule{Complete: "files"},
			wantDirective: cobra.ShellCompDirectiveDefault,
		},
		{
			name:          "files with extensions, dot stripped",
			rule:          config.Rule{Complete: "files", Extensions: []string{".go", "md"}},
			want:          []string{"go", "md"},
			wantDirective: cobra.ShellCompDirectiveFilterFileExt,
		},
		{
			name:          "urls, all when prefix empty",
			rule:          config.Rule{Complete: "urls", URLs: []string{"https://z.example", "https://a.example"}},
			want:          []string{"https://a.example", "https://z.example"},
			wantDirective: cobra.ShellCompDirectiveDefault,
		},
		{
			name:          "urls, filtered by prefix",
			rule:          config.Rule{Complete: "urls", URLs: []string{"https://github.com", "https://example.com"}},
			toComplete:    "https://g",
			want:          []string{"https://github.com"},
			wantDirective: cobra.ShellCompDirectiveDefault,
		},
		{
			name:          "no complete setting falls back to files",
			rule:          config.Rule{},
			wantDirective: cobra.ShellCompDirectiveDefault,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, directive := completeAliasTarget(tt.rule, tt.toComplete)
			if directive != tt.wantDirective {
				t.Errorf("directive = %v, want %v", directive, tt.wantDirective)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("completeAliasTarget() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCompleteGitDirs(t *testing.T) {
	root := t.TempDir()
	for _, d := range []string{"repo-a/.git", "repo-b/.git", "plain", "nested/repo-c/.git"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name       string
		toComplete string
		want       []string
	}{
		{name: "top level lists only git dirs", toComplete: root + "/", want: []string{root + "/repo-a/", root + "/repo-b/"}},
		{name: "prefix filters", toComplete: root + "/repo-a", want: []string{root + "/repo-a/"}},
		{name: "non-git prefix yields nothing", toComplete: root + "/plain", want: nil},
		{name: "descends one level", toComplete: root + "/nested/", want: []string{root + "/nested/repo-c/"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, directive := completeGitDirs(tt.toComplete)
			if directive != cobra.ShellCompDirectiveNoSpace|cobra.ShellCompDirectiveNoFileComp {
				t.Errorf("directive = %v", directive)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("completeGitDirs(%q) = %q, want %q", tt.toComplete, got, tt.want)
			}
		})
	}
}

func TestExpandUser(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if got := expandUser("~/code"); got != home+"/code" {
		t.Errorf("expandUser(~/code) = %q, want %q", got, home+"/code")
	}
	if got := expandUser("relative/path"); got != "relative/path" {
		t.Errorf("expandUser(relative/path) = %q, want unchanged", got)
	}
}

func TestCompleteArgNoConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	got, directive := completeArg(nil, nil, "z")
	if got != nil {
		t.Errorf("completeArg with no config = %q, want nil", got)
	}
	if directive != cobra.ShellCompDirectiveDefault {
		t.Errorf("directive = %v, want ShellCompDirectiveDefault", directive)
	}
}
