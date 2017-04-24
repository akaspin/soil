// +build !debug

package logx

// Debug logs value with DEBUG severity level only
// if "debug" tag is provided on build.
func (l *Log) Debug(v ...interface{}) {}

// Debugf logs formatted value with DEBUG severity level only
// if "debug" tag is provided on build.
func (l *Log) Debugf(format string, v ...interface{}) {}

func Debug(v ...interface{}) {}
func Debugf(format string, v ...interface{}) {}
