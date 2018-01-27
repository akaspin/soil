package logx

import (
	"fmt"
)

const (
	lNotice   = "NOTICE"
	lWarning  = "WARNING"
	lError    = "ERROR"
	lCritical = "CRITICAL"
)

type Log struct {
	prefix string
	tags   []string

	appender Appender
}

// Create new log
func NewLog(appender Appender, prefix string, tags ...string) (res *Log) {
	return &Log{
		tags:     tags,
		prefix:   prefix,
		appender: appender.Clone(prefix, tags),
	}
}

// NewTextAppender log with given prefix and tags.
func (l *Log) GetLog(prefix string, tags ...string) (res *Log) {
	return &Log{
		prefix:   prefix,
		tags:     tags,
		appender: l.appender.Clone(prefix, tags),
	}
}

// Log prefix.
func (l *Log) Prefix() (res string) {
	return l.prefix
}

// Log tags.
func (l *Log) Tags() (res []string) {
	return l.tags
}

// NewTextAppender log instance wit given tags
func (l *Log) WithTags(tags ...string) (res *Log) {
	return NewLog(l.appender, l.prefix, tags...)
}

// Notice logs value with NOTICE severity level.
func (l *Log) Notice(v ...interface{}) {
	l.appender.Append(lNotice, fmt.Sprint(v...))
}

// Noticef logs formatted value with NOTICE severity level.
func (l *Log) Noticef(format string, v ...interface{}) {
	l.appender.Append(lNotice, fmt.Sprintf(format, v...))
}

// Warning logs value with WARNING severity level.
func (l *Log) Warning(v ...interface{}) {
	l.appender.Append(lWarning, fmt.Sprint(v...))
}

// Warningf logs formatted value with WARNING severity level.
func (l *Log) Warningf(format string, v ...interface{}) {
	l.appender.Append(lWarning, fmt.Sprintf(format, v...))
}

// Error logs value with ERROR severity level.
func (l *Log) Error(v ...interface{}) {
	l.appender.Append(lError, fmt.Sprint(v...))
}

// Errorf logs formatted value with ERROR severity level.
func (l *Log) Errorf(format string, v ...interface{}) {
	l.appender.Append(lError, fmt.Sprintf(format, v...))
}

// Critical logs value with CRITICAL severity level.
func (l *Log) Critical(v ...interface{}) {
	l.appender.Append(lCritical, fmt.Sprint(v...))
}

// Criticalf logs formatted value with CRITICAL severity level.
func (l *Log) Criticalf(format string, v ...interface{}) {
	l.appender.Append(lCritical, fmt.Sprintf(format, v...))
}
