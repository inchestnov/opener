// Package opener resolves a CLI target into an Action describing how it
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

// Resolve decides how to launch targets, given the user's config, logging
// each decision stage via diag.
//
// alias is empty for automatic mode ("opener <target>"), where targets
// holds exactly one element. If that target is a file matching a pattern
// under cfg.Open.Patterns (checked in order, first match wins), that rule
// is used. A directory target uses cfg.Open.Directory if set. Otherwise
// resolution falls back to the system `open` command.
//
// alias is non-empty for alias mode ("opener <alias> <target>..."), where
// it is looked up in cfg.Aliases and launched as either a macOS application
// or a CLI command; an alias not present in cfg.Aliases is an error.
func Resolve(alias string, targets []string, cfg *config.Config, diag diagnostic.Context) (Action, error) {
	if alias == "" {
		return resolveAutomatic(targets, cfg, diag)
	}
	return resolveAlias(alias, targets, cfg, diag.Logger)
}

func resolveAutomatic(targets []string, cfg *config.Config, diag diagnostic.Context) (Action, error) {
	logger := diag.Logger
	target := targets[0]
	logger.Debug("target: %s", target)

	targetType := resolveTargetType(target)
	logger.Debug("target type: %s", targetType)

	switch targetType {
	case TargetFile:
		ext := extensionOf(target)
		if ext == "" {
			logger.Debug("file type: (none)")
		} else {
			logger.Debug("file type: %s", ext)
		}

		logger.Debug("checking config: %s", diag.ConfigPath)

		base := filepath.Base(target)
		for _, rule := range cfg.Open.Patterns {
			matched, err := matchPattern(rule.Pattern, base)
			if err != nil {
				return Action{}, fmt.Errorf("invalid pattern %q: %w", rule.Pattern, err)
			}
			if !matched {
				continue
			}
			logger.Debug("pattern matched: %s", rule.Pattern)
			action, err := actionForPatternRule(rule, targets)
			if err != nil {
				return Action{}, err
			}
			logCommandOrApp(logger, rule.Cmd, rule.App)
			return action, nil
		}
		logger.Debug("no pattern matched")

	case TargetDirectory:
		logger.Debug("checking config: %s", diag.ConfigPath)

		if rule := cfg.Open.Directory; rule.App != "" || rule.Cmd != "" {
			logger.Debug("config rule found: open.directory")
			logCommandOrApp(logger, rule.Cmd, rule.App)
			return actionForRule(rule, targets), nil
		}
		logger.Debug("no custom directory rule found")
		logger.Debug("using default application: Finder")
	}

	return Action{Strategy: StrategyFallback, Args: targets}, nil
}

func resolveAlias(alias string, targets []string, cfg *config.Config, logger diagnostic.Logger) (Action, error) {
	logger.Debug("alias: %s", alias)
	for _, target := range targets {
		logger.Debug("target: %s", target)
	}
	logger.Debug("checking aliases")

	rule, ok := cfg.Aliases[alias]
	if !ok {
		logger.Debug("alias not found: %s", alias)
		return Action{}, fmt.Errorf("unknown alias: %s", alias)
	}
	logger.Debug("alias found: %s", alias)
	logAliasRuleChoice(logger, rule)

	return actionForRule(rule, targets), nil
}

// logCommandOrApp logs which of a command/app pair will be used to launch
// a target. command takes precedence when both are set.
func logCommandOrApp(logger diagnostic.Logger, command, app string) {
	if command != "" {
		logger.Debug("configured command: %s", command)
		logger.Debug("launch strategy: CLI command")
		return
	}
	logger.Debug("configured application: %s", app)
	logger.Debug("launch strategy: macOS application")
}

// logAliasRuleChoice logs which of a Rule's App/Cmd forms will be used
// for an alias-mode rule.
func logAliasRuleChoice(logger diagnostic.Logger, rule config.Rule) {
	if rule.Cmd != "" {
		logger.Debug("alias type: command")
		logger.Debug("executable: %s", rule.Cmd)
		return
	}
	logger.Debug("alias type: application")
	logger.Debug("application: %s", rule.App)
}

// actionForRule builds the Action for a config.Rule against targets.
// Cmd takes precedence when both App and Cmd are set.
func actionForRule(rule config.Rule, targets []string) Action {
	if rule.Cmd != "" {
		return Action{Strategy: StrategyCommand, Name: rule.Cmd, Args: targets}
	}
	return Action{Strategy: StrategyApp, Name: rule.App, Args: targets}
}

// extensionOf returns target's file extension, lowercased and without the
// leading dot ("document.PDF" -> "pdf"; "README" -> "").
func extensionOf(target string) string {
	return strings.ToLower(strings.TrimPrefix(filepath.Ext(target), "."))
}
