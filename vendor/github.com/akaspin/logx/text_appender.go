package logx

import (
	"bytes"
	"io"
	"runtime"
	"sync"
	"time"
	"unicode"
)

const (
	// Ldate adds the date in the local time zone: 2009/01/23
	Ldate = 1 << iota

	// Ltime adds the time in the local time zone: 01:23:23
	Ltime

	// Lmicroseconds adds microsecond resolution: 01:23:23.123123. Assumes LTime.
	Lmicroseconds

	// Llongfile adds full file name and line number: /a/b/c/d.go:23
	Llongfile

	// Lshortfile adds final file name element and line number: d.go:23.
	// overrides Llongfile
	Lshortfile

	// LUTC if Ldate or Ltime is set, use UTC rather than the local time zone
	LUTC

	// Lcompact removes whitespace from log lines
	Lcompact

	// LstdFlags initial values for the standard logger
	LstdFlags = Lshortfile | Lcompact

	lCallLevel = 2
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

/*
TextAppender is default for logx

Format:

	time LEVEL prefix [tags] file:line message
*/
type TextAppender struct {
	output io.Writer
	flags  int

	identity []byte
}

// NewTextAppender returns new appender without prefix and tags
func NewTextAppender(output io.Writer, flags int) (a *TextAppender) {
	a = &TextAppender{
		output: output,
		flags:  flags,
		identity: []byte(" "),
	}
	return a
}

// Clone returns copy of TextAppender with given prefix and tags
func (a *TextAppender) Clone(prefix string, tags []string) (a1 Appender) {
	a1 = &TextAppender{
		output: a.output,
		flags:  a.flags,
	}
	a1.(*TextAppender).setIdentity(prefix, tags)
	return a1
}

func (a *TextAppender) Append(level, line string) {
	buf := bufferPool.Get().(*bytes.Buffer)

	// time
	if a.flags&(Ldate|Ltime|Lmicroseconds|LUTC) != 0 {
		t := time.Now()
		if a.flags&LUTC != 0 {
			t = t.UTC()
		}
		if a.flags&(Ldate|Ltime|Lmicroseconds) != 0 {
			if a.flags&Ldate != 0 {
				year, month, day := t.Date()
				itoaBuf(buf, year, 4)
				buf.WriteByte('/')
				itoaBuf(buf, int(month), 2)
				buf.WriteByte('/')
				itoaBuf(buf, day, 2)
				buf.WriteByte(' ')
			}
			if a.flags&(Ltime|Lmicroseconds) != 0 {
				hour, min, sec := t.Clock()
				itoaBuf(buf, hour, 2)
				buf.WriteByte(':')
				itoaBuf(buf, min, 2)
				buf.WriteByte(':')
				itoaBuf(buf, sec, 2)
				if a.flags&Lmicroseconds != 0 {
					buf.WriteByte('.')
					itoaBuf(buf, t.Nanosecond()/1e3, 6)
				}
				buf.WriteByte(' ')
			}
		}
	}

	// level
	buf.WriteString(level)

	// identity
	buf.Write(a.identity)

	// file
	if a.flags&(Lshortfile|Llongfile) != 0 {
		_, file, lineNo, ok := runtime.Caller(lCallLevel)
		if !ok {
			file = "???"
			lineNo = 0
		}
		if a.flags&Lshortfile != 0 {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}
		buf.WriteString(file)
		buf.WriteByte(':')
		itoaBuf(buf, lineNo, -1)
		buf.WriteByte(' ')
	}

	if a.flags&(Lcompact) != 0 {
		stripBuf(buf, line)
	} else {
		buf.WriteString(line)
	}

	if len(line) == 0 || line[len(line)-1] != '\n' {
		buf.WriteByte('\n')
	}
	buf.WriteTo(a.output)
	buf.Reset()
	bufferPool.Put(buf)
}

func (a *TextAppender) setIdentity(prefix string, tags []string) {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.WriteByte(' ')
	if prefix != "" {
		buf.WriteString(prefix)
		buf.WriteByte(' ')
	}
	if len(tags) > 0 {
		buf.WriteByte('[')
		for i, tag := range tags {
			if i > 0 {
				buf.WriteByte(' ')
			}
			buf.WriteString(tag)
		}
		buf.WriteByte(']')
		buf.WriteByte(' ')
	}
	a.identity = make([]byte, buf.Len())
	copy(a.identity, buf.Bytes())
	buf.Reset()
	bufferPool.Put(buf)
}

func itoaBuf(buf *bytes.Buffer, i int, wid int) {
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	buf.Write(b[bp:])
}

func stripBuf(buf *bytes.Buffer, src string) {
	var white bool
	for _, c := range src {
		if unicode.IsSpace(c) {
			if !white {
				buf.WriteByte(' ')
			}
			white = true
		} else {
			buf.WriteRune(c)
			white = false
		}
	}
}
