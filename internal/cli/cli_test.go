package cli

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/spf13/cobra"
)

func TestCompleteArg(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	const cfg = `
aliases:
  code:
    app: "Visual Studio Code"
  web:
    app: "Google Chrome"
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
			name:       "second arg: no name candidates, file completion only",
			args:       []string{"code"},
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
