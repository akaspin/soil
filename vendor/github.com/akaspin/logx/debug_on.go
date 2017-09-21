// +build debug trace

package logx

import (
	"fmt"
)

const (
	lDebug = "DEBUG"
)

// Debug logs value with DEBUG severity level only
// if "debug" tag is provided on build.
func (l *Log) Debug(v ...interface{}) {
	l.appendLine(lDebug, fmt.Sprint(v...))
}

// Debugf logs formatted value with DEBUG severity level only
// if "debug" tag is provided on build.
func (l *Log) Debugf(format string, v ...interface{}) {
	l.appendLine(lDebug,  fmt.Sprintf(format, v...))
}

// Debug logs value with DEBUG severity level only
// if "debug" tag is provided on build.
func Debug(v ...interface{}) {
	std.appendLine(lDebug,  fmt.Sprint(v...))
}

// Debugf logs formatted value with DEBUG severity level only
// if "debug" tag is provided on build.
func Debugf(format string, v ...interface{}) {
	std.appendLine(lDebug,  fmt.Sprintf(format, v...))
}
