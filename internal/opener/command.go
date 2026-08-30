package opener

import (
	"fmt"
	"strings"
	"unicode"
)

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
