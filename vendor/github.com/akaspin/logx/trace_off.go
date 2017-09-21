// +build !trace

package logx

// Trace logs value with TRACE severity level only
// if "trace" tag is provided on build.
func (l *Log) Trace(v ...interface{}) {}

// Tracef logs formatted value with TRACE severity level only
// if "trace" tag is provided on build.
func (l *Log) Tracef(format string, v ...interface{}) {}

// Trace logs value with TRACE severity level only
// if "trace" tag is provided on build.
func Trace(v ...interface{}) {}

// Tracef logs formatted value with TRACE severity level only
// if "trace" tag is provided on build.
func Tracef(format string, v ...interface{}) {}
