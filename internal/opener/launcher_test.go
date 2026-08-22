package opener

import (
	"os/exec"
	"testing"
)

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
