package logx

import (
	"os"
)

// Default TextAppender
var DefaultAppender = NewTextAppender(os.Stderr, LstdFlags)

var std = NewLog(DefaultAppender, "")

// GetLog returns new independent log instance with given prefix
func GetLog(prefix string, tags ...string) *Log {
	return std.GetLog(prefix, tags...)
}
