// +build trace

package logx

import (
	"fmt"
)

const (
	lTrace = "TRACE"
)

// Trace logs value with TRACE severity level only
// if "trace" tag is provided on build.
func (l *Log) Trace(v ...interface{}) {
	l.appendLine(lTrace, fmt.Sprint(v...))
}

// Tracef logs formatted value with TRACE severity level only
// if "trace" tag is provided on build.
func (l *Log) Tracef(format string, v ...interface{}) {
	l.appendLine(lTrace, fmt.Sprintf(format, v...))
}

// Trace logs value with TRACE severity level only
// if "trace" tag is provided on build.
func Trace(v ...interface{}) {
	std.appendLine(lTrace, fmt.Sprint(v...))
}

// Tracef logs formatted value with TRACE severity level only
// if "trace" tag is provided on build.
func Tracef(format string, v ...interface{}) {
	std.appendLine(lTrace, fmt.Sprintf(format, v...))
}
