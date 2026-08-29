package opener

import (
	"errors"
	"os/exec"
	"strings"
	"testing"
)

func TestOpenError(t *testing.T) {
	const target = "go.mod"

	tests := []struct {
		name       string
		action     Action
		stderr     string
		wantSubstr []string
		wantNoRaw  bool
	}{
		{
			name:   "no app registered for file type",
			action: Action{Strategy: StrategyFallback, Args: []string{target}},
			stderr: `No application knows how to open URL file:///Users/x/go.mod ` +
				`(Error Domain=NSOSStatusErrorDomain Code=-10814 "kLSApplicationNotFoundErr").`,
			wantSubstr: []string{"nothing knows how to open go.mod", "open.patterns", "Finder"},
			wantNoRaw:  true,
		},
		{
			name:       "app strategy, app not installed",
			action:     Action{Strategy: StrategyApp, Name: "Sublime Text", Args: []string{target}},
			stderr:     "Unable to find application named 'Sublime Text' (-10810)",
			wantSubstr: []string{`"Sublime Text" is not installed`, "~/.opener.yaml"},
			wantNoRaw:  true,
		},
		{
			name:       "unclassified stderr is surfaced, first line only",
			action:     Action{Strategy: StrategyFallback, Args: []string{target}},
			stderr:     "the file is on fire\nsecond line\nthird line",
			wantSubstr: []string{"could not open go.mod", "the file is on fire"},
		},
		{
			name:       "empty stderr falls back to run error",
			action:     Action{Strategy: StrategyFallback, Args: []string{target}},
			stderr:     "",
			wantSubstr: []string{"could not open go.mod", "boom"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := openError(tt.action, tt.stderr, errors.New("boom"))
			if err == nil {
				t.Fatal("openError() = nil, want error")
			}
			got := err.Error()
			for _, want := range tt.wantSubstr {
				if !strings.Contains(got, want) {
					t.Errorf("openError() = %q, missing %q", got, want)
				}
			}
			if tt.wantNoRaw {
				for _, noise := range []string{"NSOSStatusErrorDomain", "file://", "-10814", "-10810"} {
					if strings.Contains(got, noise) {
						t.Errorf("openError() = %q, leaked raw noise %q", got, noise)
					}
				}
			}
		})
	}
}

func TestFormatCommand(t *testing.T) {
	tests := []struct {
		name string
		cmd  *exec.Cmd
		want string
	}{
		{
			name: "no args needing quotes",
			cmd:  exec.Command("open", "./project"),
			want: "open ./project",
		},
		{
			name: "app name with a space is quoted",
			cmd:  exec.Command("open", "-a", "Google Chrome", "document.pdf"),
			want: `open -a "Google Chrome" document.pdf`,
		},
		{
			name: "app name without a space is not quoted",
			cmd:  exec.Command("open", "-a", "TextEdit", "."),
			want: "open -a TextEdit .",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatCommand(tt.cmd); got != tt.want {
				t.Errorf("formatCommand() = %q, want %q", got, tt.want)
			}
		})
	}
}
