// Package diagnostic provides opener's verbose-mode diagnostic logging
// interface, shared across resolution and launching.
package diagnostic

import (
	"fmt"
	"io"
)

// Logger receives diagnostic trace messages describing resolution
// decisions. Debug is a no-op unless verbose mode is enabled.
type Logger interface {
	Debug(format string, args ...any)
}

type noopLogger struct{}

func (noopLogger) Debug(string, ...any) {}

// Noop discards all diagnostic messages.
var Noop Logger = noopLogger{}

type writerLogger struct {
	w io.Writer
}

// NewWriterLogger returns a Logger that writes each Debug call as a
// "[verbose] "-prefixed line to w.
func NewWriterLogger(w io.Writer) Logger {
	return writerLogger{w: w}
}

func (l writerLogger) Debug(format string, args ...any) {
	fmt.Fprintf(l.w, "[verbose] "+format+"\n", args...)
}
