// +build !trace

package logx

// Trace logs value with TRACE severity level only
// if "trace" tag is provided on build.
func (*Log) Trace(v ...interface{}) {}

// Tracef logs formatted value with TRACE severity level only
// if "trace" tag is provided on build.
func (*Log) Tracef(format string, v ...interface{}) {}
