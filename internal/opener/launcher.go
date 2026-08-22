package opener

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/inchestnov/opener/internal/diagnostic"
)

// Launch executes the resolved Action. Commands are always run directly via
// os/exec, never through a shell.
func Launch(action Action, logger diagnostic.Logger) error {
	var cmd *exec.Cmd

	switch action.Strategy {
	case StrategyApp:
		args := append([]string{"-a", action.Name}, action.Args...)
		cmd = exec.Command("open", args...)
	case StrategyCommand:
		cmd = exec.Command(action.Name, action.Args...)
	case StrategyFallback:
		cmd = exec.Command("open", action.Args...)
	default:
		return fmt.Errorf("unknown launch strategy: %v", action.Strategy)
	}

	rendered := formatCommand(cmd)
	logger.Debug("command: %s", rendered)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("launching %s: %w", rendered, err)
	}
	return nil
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
