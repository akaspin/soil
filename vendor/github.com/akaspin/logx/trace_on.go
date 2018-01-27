// +build trace,!notice

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
	l.appender.Append(lTrace, fmt.Sprint(v...))
}

// Tracef logs formatted value with TRACE severity level only
// if "trace" tag is provided on build.
func (l *Log) Tracef(format string, v ...interface{}) {
	l.appender.Append(lTrace, fmt.Sprintf(format, v...))
}
