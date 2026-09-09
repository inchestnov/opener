// Package opener resolves a CLI invocation into an Action describing how it
// should be launched, then executes that Action via the launcher.
package opener

import (
	"fmt"
	"path/filepath"
	"strings"

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
// resulting argv. An alias's source, if any, is not consulted here.
//
// Targets are passed through exactly as typed unless the alias sets a
// `base:`, in which case each target that is not already anchored
// elsewhere is joined onto it - see rebaseTarget.
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

	if a.Base != "" {
		logger.Debug("base: %s", a.Base)
		targets = rebaseTargets(targets, a.Base)
		for _, target := range targets {
			logger.Debug("rebased target: %s", target)
		}
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

// rebaseTargets joins each target onto base, the inverse of the trimming
// the completion side does, so that what you tab-complete is what opens.
func rebaseTargets(targets []string, base string) []string {
	out := make([]string, len(targets))
	for i, target := range targets {
		out[i] = rebaseTarget(target, base)
	}
	return out
}

// rebaseTarget joins target onto base, unless target already says where it
// lives: an absolute path, a ~ path, an explicit ./ or ../ path, or
// anything carrying a URL scheme. Those escape hatches are what let an
// alias with a base still open a file from somewhere else entirely.
func rebaseTarget(target, base string) string {
	switch {
	case target == "":
		return target
	case filepath.IsAbs(target):
		return target
	case target == "~" || strings.HasPrefix(target, "~/"):
		return target
	case target == "." || target == "..":
		return target
	case strings.HasPrefix(target, "./") || strings.HasPrefix(target, "../"):
		return target
	case strings.Contains(target, "://"):
		return target
	default:
		return filepath.Join(base, target)
	}
}
