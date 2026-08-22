// Package opener resolves a CLI target into an Action describing how it
// should be launched, then executes that Action via the launcher.
package opener

import "github.com/inchestnov/opener/internal/config"

// Strategy is how a resolved Action should be launched.
type Strategy int

const (
	// StrategyFallback delegates entirely to the system `open` command.
	StrategyFallback Strategy = iota
	// StrategyApp launches a macOS GUI application via `open -a`.
	StrategyApp
	// StrategyCommand runs a CLI executable directly.
	StrategyCommand
)

// Action is the resolved decision of how to launch a target: which
// strategy to use, the app name or executable (empty for StrategyFallback),
// and the final arguments to pass to it.
type Action struct {
	Strategy Strategy
	Name     string
	Args     []string
}

// Resolve decides how target should be opened, given the user's config.
// In the absence of a matching rule, it falls back to the system `open`
// command for both files and directories, and for targets that don't
// resolve to a local path (URLs, nonexistent paths).
func Resolve(target string, cfg *config.Config) (Action, error) {
	return Action{
		Strategy: StrategyFallback,
		Args:     []string{target},
	}, nil
}
