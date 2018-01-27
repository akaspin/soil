// +build !debug,!trace

package logx

// Debug logs value with DEBUG severity level only
// if "debug" tag is provided on build.
func (*Log) Debug(v ...interface{}) {}

// Debugf logs formatted value with DEBUG severity level only
// if "debug" tag is provided on build.
func (*Log) Debugf(format string, v ...interface{}) {}
