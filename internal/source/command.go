package source

import (
	"context"
	"io"
	"os/exec"
	"strings"
	"time"
)

// commandTimeout bounds a command source: it runs on every <TAB>, so a slow
// or hung command yields nothing rather than stalling the shell.
const commandTimeout = 2 * time.Second

// commandSource runs a shell command and treats each line of its stdout as
// a candidate. The command is run via `sh -c` so pipes, globs, and $HOME
// work; it comes from the user's own ~/.opener.yaml.
type commandSource struct {
	run string
	cwd string
}

func (c *commandSource) Candidates(toComplete string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", c.run)
	if c.cwd != "" {
		cmd.Dir = expandUser(c.cwd)
	}
	cmd.Stderr = io.Discard

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return filterSort(strings.Split(string(out), "\n"), toComplete), nil
}
