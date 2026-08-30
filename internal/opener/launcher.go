package opener

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/inchestnov/opener/internal/diagnostic"
)

// Launch executes the resolved Action. Commands are always run directly via
// os/exec, never through a shell.
func Launch(action Action, logger diagnostic.Logger) error {
	var cmd *exec.Cmd
	viaOpen := false

	switch action.Strategy {
	case StrategyApp:
		args := append([]string{"-a", action.Name}, action.Args...)
		cmd = exec.Command("open", args...)
		viaOpen = true
	case StrategyCommand:
		cmd = exec.Command(action.Name, action.Args...)
		// An alias of `cmd: open` is still `open` under the hood: capture
		// its stderr and translate LaunchServices failures like we do for
		// StrategyApp.
		viaOpen = filepath.Base(action.Name) == "open"
	default:
		return fmt.Errorf("unknown launch strategy: %v", action.Strategy)
	}

	rendered := formatCommand(cmd)
	logger.Debug("command: %s", rendered)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	// A CLI command may be an interactive editor: leave its stderr wired to
	// ours. `open` is non-interactive, so capture its stderr and translate
	// its LaunchServices failures into something actionable.
	if !viaOpen {
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("launching %s: %w", rendered, err)
		}
		return nil
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if raw := strings.TrimSpace(stderr.String()); raw != "" {
		logger.Debug("open stderr: %s", raw)
	}
	if err != nil {
		return openError(action, stderr.String(), err)
	}
	return nil
}

// openError turns a failed `open` invocation into an actionable message,
// classifying its stderr rather than surfacing raw LaunchServices noise or a
// bare "exit status 1".
func openError(action Action, stderr string, runErr error) error {
	msg := strings.TrimSpace(stderr)
	targets := strings.Join(action.Args, " ")
	low := strings.ToLower(msg)

	switch {
	// -10810 kLSApplicationNotFoundErr for `open -a Name`: the app itself
	// is missing.
	case action.Strategy == StrategyApp &&
		(strings.Contains(msg, "-10810") ||
			strings.Contains(low, "unable to find application")):
		return fmt.Errorf("application %q is not installed\n"+
			"  it is referenced by an alias in ~/.opener.yaml", action.Name)

	// -10814 kLSApplicationNotFoundErr for `open <file>`: no app is
	// registered for this file type.
	case strings.Contains(msg, "-10814"),
		strings.Contains(msg, "kLSApplicationNotFoundErr"),
		strings.Contains(low, "no application knows how to open"):
		return fmt.Errorf("nothing knows how to open %s\n"+
			"  no app is associated with this file type. Fix it by either:\n"+
			"    - adding an alias with an explicit app/cmd in ~/.opener.yaml\n"+
			"    - setting a default app in Finder: Get Info ▸ Open with ▸ Change All", targets)

	case msg != "":
		return fmt.Errorf("could not open %s: %s", targets, firstLine(msg))

	default:
		return fmt.Errorf("could not open %s: %w", targets, runErr)
	}
}

// firstLine returns s up to its first newline, trimmed.
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}

// formatCommand renders cmd's argv for diagnostic output, quoting any
// argument containing whitespace.
func formatCommand(cmd *exec.Cmd) string {
	parts := make([]string, len(cmd.Args))
	for i, arg := range cmd.Args {
		if strings.ContainsAny(arg, " \t") {
			parts[i] = fmt.Sprintf("%q", arg)
		} else {
			parts[i] = arg
		}
	}
	return strings.Join(parts, " ")
}
