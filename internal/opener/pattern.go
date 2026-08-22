package opener

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/inchestnov/opener/internal/config"
)

// matchPattern reports whether filename matches pattern, case-insensitively.
// A bare extension pattern (starts with "." and has no glob metacharacters,
// e.g. ".pdf") is treated as "*.pdf". Anything else is matched via
// filepath.Match's glob syntax (*, ?, [...]).
func matchPattern(pattern, filename string) (bool, error) {
	if strings.HasPrefix(pattern, ".") && !strings.ContainsAny(pattern, "*?[") {
		pattern = "*" + pattern
	}
	return filepath.Match(strings.ToLower(pattern), strings.ToLower(filename))
}

// actionForPatternRule builds the Action for a config.PatternRule against
// targets. Cmd takes precedence when both App and Cmd are set: it is
// shell-word-split (never run through an actual shell) and targets are
// appended to the resulting argv.
func actionForPatternRule(rule config.PatternRule, targets []string) (Action, error) {
	if rule.Cmd != "" {
		argv, err := splitCommand(rule.Cmd)
		if err != nil {
			return Action{}, fmt.Errorf("parsing cmd for pattern %q: %w", rule.Pattern, err)
		}
		if len(argv) == 0 {
			return Action{}, fmt.Errorf("empty cmd for pattern %q", rule.Pattern)
		}
		return Action{Strategy: StrategyCommand, Name: argv[0], Args: append(argv[1:], targets...)}, nil
	}
	return Action{Strategy: StrategyApp, Name: rule.App, Args: targets}, nil
}

// splitCommand splits s into words the way a shell would, honoring single-
// and double-quoted substrings, without invoking an actual shell.
func splitCommand(s string) ([]string, error) {
	var args []string
	var buf strings.Builder
	var quote rune
	inArg := false

	for _, r := range s {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				buf.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote = r
			inArg = true
		case unicode.IsSpace(r):
			if inArg {
				args = append(args, buf.String())
				buf.Reset()
				inArg = false
			}
		default:
			buf.WriteRune(r)
			inArg = true
		}
	}
	if quote != 0 {
		return nil, fmt.Errorf("unterminated %c quote", quote)
	}
	if inArg {
		args = append(args, buf.String())
	}
	return args, nil
}
