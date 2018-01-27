// +build debug,!notice trace,!notice

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
	l.appender.Append(lDebug, fmt.Sprint(v...))
}

// Debugf logs formatted value with DEBUG severity level only
// if "debug" tag is provided on build.
func (l *Log) Debugf(format string, v ...interface{}) {
	l.appender.Append(lDebug, fmt.Sprintf(format, v...))
}
