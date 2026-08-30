// Package opener resolves a CLI invocation into an Action describing how it
// should be launched, then executes that Action via the launcher.
package opener

import (
	"fmt"

	"github.com/inchestnov/opener/internal/config"
	"github.com/inchestnov/opener/internal/diagnostic"
)

// Strategy is how a resolved Action should be launched.
type Strategy int

const (
	// StrategyApp launches a macOS GUI application via `open -a`.
	StrategyApp Strategy = iota
	// StrategyCommand runs a CLI executable directly.
	StrategyCommand
)

// Action is the resolved decision of how to launch targets: which strategy
// to use, the app name or executable, and the arguments to pass to it.
type Action struct {
	Strategy Strategy
	Name     string
	Args     []string
}

// Resolve looks alias up in cfg.Aliases and builds the Action for launching
// targets with it. An alias not present in cfg.Aliases is an error, as is
// one with neither app nor cmd set.
//
// An alias's cmd is split into words the way a shell would (quotes
// honored), never run through a shell, with targets appended to the
// resulting argv. An alias's source, if any, is not consulted here: targets
// are passed through exactly as typed.
func Resolve(alias string, targets []string, cfg *config.Config, logger diagnostic.Logger) (Action, error) {
	logger.Debug("alias: %s", alias)
	for _, target := range targets {
		logger.Debug("target: %s", target)
	}

	a, ok := cfg.Aliases[alias]
	if !ok {
		logger.Debug("alias not found: %s", alias)
		return Action{}, fmt.Errorf("unknown alias: %s", alias)
	}
	if a.App == "" && a.Cmd == "" {
		return Action{}, fmt.Errorf("alias %q has neither app nor cmd configured", alias)
	}

	if a.Cmd != "" {
		argv, err := splitCommand(a.Cmd)
		if err != nil {
			return Action{}, fmt.Errorf("parsing cmd for alias %q: %w", alias, err)
		}
		if len(argv) == 0 {
			return Action{}, fmt.Errorf("empty cmd for alias %q", alias)
		}
		logger.Debug("alias type: command")
		logger.Debug("executable: %s", argv[0])
		return Action{Strategy: StrategyCommand, Name: argv[0], Args: append(argv[1:], targets...)}, nil
	}

	logger.Debug("alias type: application")
	logger.Debug("application: %s", a.App)
	return Action{Strategy: StrategyApp, Name: a.App, Args: targets}, nil
}
