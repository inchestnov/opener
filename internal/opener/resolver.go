// Package opener resolves a CLI target into an Action describing how it
// should be launched, then executes that Action via the launcher.
package opener

import (
	"fmt"

	"github.com/inchestnov/opener/internal/config"
)

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

// Resolve decides how to launch targets, given the user's config.
//
// alias is empty for automatic mode ("opener <target>"), where targets
// holds exactly one element and resolution falls back to the system `open`
// command in the absence of a matching rule.
//
// alias is non-empty for alias mode ("opener <alias> <target>..."), where
// it is looked up in cfg.Aliases and launched as either a macOS application
// or a CLI command; an alias not present in cfg.Aliases is an error.
func Resolve(alias string, targets []string, cfg *config.Config) (Action, error) {
	if alias == "" {
		return Action{
			Strategy: StrategyFallback,
			Args:     targets,
		}, nil
	}

	rule, ok := cfg.Aliases[alias]
	if !ok {
		return Action{}, fmt.Errorf("unknown alias: %s", alias)
	}

	return actionForRule(rule, targets), nil
}

// actionForRule builds the Action for a config.Rule against targets.
// Command takes precedence when both App and Command are set.
func actionForRule(rule config.Rule, targets []string) Action {
	if rule.Command != "" {
		return Action{Strategy: StrategyCommand, Name: rule.Command, Args: targets}
	}
	return Action{Strategy: StrategyApp, Name: rule.App, Args: targets}
}
