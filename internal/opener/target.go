package opener

import "os"

// TargetType classifies what kind of thing a target argument points at.
type TargetType int

const (
	TargetUnknown TargetType = iota
	TargetFile
	TargetDirectory
)

// String renders t for diagnostic output.
func (t TargetType) String() string {
	switch t {
	case TargetFile:
		return "file"
	case TargetDirectory:
		return "directory"
	default:
		return "unknown"
	}
}

// resolveTargetType classifies target by stat-ing it. A target that cannot
// be stat-ed (a URL, a nonexistent path) is TargetUnknown and is left to the
// system open command to interpret.
func resolveTargetType(target string) TargetType {
	info, err := os.Stat(target)
	if err != nil {
		return TargetUnknown
	}
	if info.IsDir() {
		return TargetDirectory
	}
	return TargetFile
}
