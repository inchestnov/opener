package opener

import (
	"fmt"
	"os"
	"os/exec"
)

// Launch executes the resolved Action. Commands are always run directly via
// os/exec, never through a shell.
func Launch(action Action) error {
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

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
