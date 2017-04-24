package logx


type Appender interface {

	// Append log line
	Append(level, prefix, line string, tags ...string)
}
