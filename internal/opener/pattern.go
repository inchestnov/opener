package opener

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/inchestnov/opener/internal/config"
)

// matchPattern reports whether filename matches pattern, case-insensitively.
// A brace group (e.g. "*.{jpg,png}") expands into one alternative per
// comma-separated term; filename matches if any alternative does. Within
// each alternative, a bare extension (starts with "." and has no glob
// metacharacters, e.g. ".pdf") is treated as "*.pdf". Anything else is
// matched via filepath.Match's glob syntax (*, ?, [...]).
func matchPattern(pattern, filename string) (bool, error) {
	filename = strings.ToLower(filename)
	for _, alt := range expandBraces(strings.ToLower(pattern)) {
		if strings.HasPrefix(alt, ".") && !strings.ContainsAny(alt, "*?[") {
			alt = "*" + alt
		}
		matched, err := filepath.Match(alt, filename)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

// expandBraces expands the first "{a,b,c}" group in pattern into one string
// per comma-separated term, recursing to expand any further groups in the
// remainder. A pattern with no brace group expands to itself.
func expandBraces(pattern string) []string {
	start := strings.IndexByte(pattern, '{')
	if start == -1 {
		return []string{pattern}
	}
	end := strings.IndexByte(pattern[start:], '}')
	if end == -1 {
		return []string{pattern}
	}
	end += start

	prefix, terms, suffix := pattern[:start], strings.Split(pattern[start+1:end], ","), pattern[end+1:]
	var out []string
	for _, term := range terms {
		for _, rest := range expandBraces(suffix) {
			out = append(out, prefix+term+rest)
		}
	}
	return out
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
