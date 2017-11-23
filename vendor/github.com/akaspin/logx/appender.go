package logx

type Appender interface {

	// Append log line. Append should be thread-safe.
	Append(level, prefix, line string, tags ...string)
}
