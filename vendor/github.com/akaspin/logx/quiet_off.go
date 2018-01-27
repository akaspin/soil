// +build !notice

package logx

import "fmt"

const lInfo = "INFO"

// Print is synonym to Info used for compatibility with "log" package.
func (l *Log) Print(v ...interface{}) {
	l.appender.Append(lInfo, fmt.Sprint(v...))
}

// Printf is synonym to Infof used for compatibility  "log" package.
func (l *Log) Printf(format string, v ...interface{}) {
	l.appender.Append(lInfo, fmt.Sprintf(format, v...))
}

// Info logs value with INFO severity level.
func (l *Log) Info(v ...interface{}) {
	l.appender.Append(lInfo, fmt.Sprint(v...))
}

// Infof logs formatted value with INFO severity level.
func (l *Log) Infof(format string, v ...interface{}) {
	l.appender.Append(lInfo, fmt.Sprintf(format, v...))
}
