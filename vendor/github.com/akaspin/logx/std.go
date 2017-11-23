package logx

import (
	"os"
)

var std = NewLog(NewSimpleAppender(os.Stderr, LstdFlags), "")

// GetLog returns new independent log instance with given prefix
func GetLog(prefix string, tags ...string) *Log {
	return std.GetLog(prefix, tags...)
}
