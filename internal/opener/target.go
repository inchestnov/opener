package opener

import (
	"os"
	"regexp"
)

// TargetType classifies what kind of thing a target argument points at.
type TargetType int

const (
	TargetUnknown TargetType = iota
	TargetFile
	TargetDirectory
	TargetExecutable
	TargetURL
)

// String renders t for diagnostic output.
func (t TargetType) String() string {
	switch t {
	case TargetFile:
		return "file"
	case TargetDirectory:
		return "directory"
	case TargetExecutable:
		return "executable"
	case TargetURL:
		return "url"
	default:
		return "unknown"
	}
}

// schemePattern matches a leading URI scheme (RFC 3986: a letter followed by
// letters, digits, "+", "-", or "."), e.g. "https:" or "mailto:".
var schemePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*:`)

// resolveTargetType classifies target by stat-ing it. A target that cannot
// be stat-ed is TargetURL if it looks like a URI (e.g. "https://...",
// "mailto:...") and TargetUnknown otherwise. An existing file with any
// executable permission bit set is TargetExecutable rather than TargetFile.
func resolveTargetType(target string) TargetType {
	info, err := os.Stat(target)
	if err != nil {
		if schemePattern.MatchString(target) {
			return TargetURL
		}
		return TargetUnknown
	}
	if info.IsDir() {
		return TargetDirectory
	}
	if info.Mode().Perm()&0o111 != 0 {
		return TargetExecutable
	}
	return TargetFile
}
