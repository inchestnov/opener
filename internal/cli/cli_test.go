package cli

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/spf13/cobra"
)

const completionCfg = `
sources:
  bookmarks:
    kind: list
    items:
      - https://github.com
      - https://example.com
aliases:
  code:
    app: "Visual Studio Code"
    source:
      kind: list
      items: ["/repos/a", "/repos/b"]
  web:
    app: "Google Chrome"
    source: bookmarks
  edit:
    cmd: nvim
  raw:
    cmd: cat
    source: does-not-exist
`

func writeCompletionConfig(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(filepath.Join(home, ".opener.yaml"), []byte(completionCfg), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCompleteArg_AliasNames(t *testing.T) {
	writeCompletionConfig(t)

	tests := []struct {
		name       string
		toComplete string
		want       []string
	}{
		{"empty prefix lists all", "", []string{"code\tapp: Visual Studio Code", "edit\tcmd: nvim", "raw\tcmd: cat", "web\tapp: Google Chrome"}},
		{"prefix filters", "e", []string{"edit\tcmd: nvim"}},
		{"no match", "zzz", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, directive := completeArg(nil, nil, tt.toComplete)
			if directive != cobra.ShellCompDirectiveNoFileComp {
				t.Errorf("directive = %v, want NoFileComp", directive)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("completeArg(nil, %q) = %q, want %q", tt.toComplete, got, tt.want)
			}
		})
	}
}

func TestCompleteArg_Targets(t *testing.T) {
	writeCompletionConfig(t)

	tests := []struct {
		name          string
		args          []string
		toComplete    string
		want          []string
		wantDirective cobra.ShellCompDirective
	}{
		{
			name:          "list source, all",
			args:          []string{"code"},
			want:          []string{"/repos/a", "/repos/b"},
			wantDirective: cobra.ShellCompDirectiveNoFileComp,
		},
		{
			name:          "named list source, prefix filtered",
			args:          []string{"web"},
			toComplete:    "https://g",
			want:          []string{"https://github.com"},
			wantDirective: cobra.ShellCompDirectiveNoFileComp,
		},
		{
			name:          "alias with no source falls back to files",
			args:          []string{"edit"},
			want:          nil,
			wantDirective: cobra.ShellCompDirectiveDefault,
		},
		{
			name:          "unknown alias offers nothing",
			args:          []string{"nope"},
			want:          nil,
			wantDirective: cobra.ShellCompDirectiveNoFileComp,
		},
		{
			name:          "broken source ref offers nothing",
			args:          []string{"raw"},
			want:          nil,
			wantDirective: cobra.ShellCompDirectiveNoFileComp,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, directive := completeArg(nil, tt.args, tt.toComplete)
			if directive != tt.wantDirective {
				t.Errorf("directive = %v, want %v", directive, tt.wantDirective)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("completeArg(%q, %q) = %q, want %q", tt.args, tt.toComplete, got, tt.want)
			}
		})
	}
}

func TestCompleteArg_NoConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	got, directive := completeArg(nil, nil, "z")
	if got != nil {
		t.Errorf("completeArg with no config = %q, want nil", got)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("directive = %v, want NoFileComp", directive)
	}
}

func TestRequireAliasAndTarget(t *testing.T) {
	for _, n := range []int{0, 1} {
		if err := requireAliasAndTarget(nil, make([]string, n)); err == nil {
			t.Errorf("requireAliasAndTarget(%d args) = nil, want error", n)
		}
	}
	if err := requireAliasAndTarget(nil, []string{"code", "."}); err != nil {
		t.Errorf("requireAliasAndTarget(2 args) = %v, want nil", err)
	}
}
