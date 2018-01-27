package logx

// Appender accepts log entries
type Appender interface {

	// Append sends log line to appender. Append should be thread-safe.
	Append(level, line string)

	// Clone returns new appender with given prefix and tags.
	Clone(prefix string, tags []string) Appender
}
