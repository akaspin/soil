// +build notice,!debug,!trace

package logx

// Print is synonym to Info used for compatibility with "log" package.
func (*Log) Print(v ...interface{}) {}

// Printf is synonym to Infof used for compatibility  "log" package.
func (*Log) Printf(format string, v ...interface{}) {}

// Info logs value with INFO severity level.
func (*Log) Info(v ...interface{}) {}

// Infof logs formatted value with INFO severity level.
func (*Log) Infof(format string, v ...interface{}) {}
