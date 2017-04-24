package logx

import (
	"os"
	"fmt"
)

var std = NewLog(NewSimpleAppender(os.Stderr, LstdFlags), "")

//
func SetAppender(appender Appender) {
	std.SetAppender(appender)
}

// GetLog returns new independent log instance with given prefix
func GetLog(prefix string) *Log {
	return std.GetLog(prefix)
}

// Print is synonym to standard Log Info used for compatibility.
func Print(v ...interface{}) {
	std.append(lInfo, fmt.Sprint(v...))
}

// Printf is synonym to standard Log Infof used for compatibility.
func Printf(format string, v ...interface{}) {
	std.append(lInfo, fmt.Sprintf(format, v...))
}

// Info logs value with INFO severity level to standard Log.
func Info(v ...interface{}) {
	std.append(lInfo, fmt.Sprint(v...))
}

// Infof logs formatted value with INFO severity level to standard Log.
func Infof(format string, v ...interface{}) {
	std.append(lInfo, fmt.Sprintf(format, v...))
}

// Warning logs value with WARNING severity level to standard Log.
func Warning(v ...interface{}) {
	std.append(lWarning, fmt.Sprint(v...))
}

// Warningf logs formatted value with WARNING severity level to standard Log.
func Warningf(format string, v ...interface{}) {
	std.append(lWarning, fmt.Sprintf(format, v...))
}

// Error logs value with ERROR severity level to standard Log.
func Error(v ...interface{}) {
	std.append(lError, fmt.Sprint(v...))
}

// Errorf logs formatted value with ERROR severity level to standard Log.
func Errorf(format string, v ...interface{}) {
	std.append(lError, fmt.Sprintf(format, v...))
}

// Critical logs value with CRITICAL severity level to standard Log.
func Critical(v ...interface{}) {
	std.append(lCritical, fmt.Sprint(v...))
}

// Criticalf logs formatted value with CRITICAL severity level to standard Log.
func Criticalf(format string, v ...interface{}) {
	std.append(lCritical, fmt.Sprintf(format, v...))
}
