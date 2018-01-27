package logx_test

// without flags

import (
	"bytes"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/stretchr/testify/assert"
	"testing"
)

func checkLevels(t *testing.T, expect ...string) {
	t.Helper()
	var w bytes.Buffer
	app := logx.NewTextAppender(&w, 0)
	l1 := logx.NewLog(app, "test")

	in := "test"
	l1.Trace(in)
	l1.Tracef("f:%s", in)
	l1.Debug(in)
	l1.Debugf("f:%s", in)
	l1.Info(in)
	l1.Infof("f:%s", in)
	l1.Notice(in)
	l1.Noticef("f:%s", in)
	l1.Warning(in)
	l1.Warningf("f:%s", in)
	l1.Error(in)
	l1.Errorf("f:%s", in)
	l1.Critical(in)
	l1.Criticalf("f:%s", in)

	var res string
	for _, level := range expect {
		res += fmt.Sprintf("%s test %s\n", level, in)
		res += fmt.Sprintf("%s test f:%s\n", level, in)
	}
	assert.Equal(t, res, w.String())
}

func TestLog_Lshortfile(t *testing.T) {
	var buf bytes.Buffer
	l := logx.NewLog(logx.NewTextAppender(&buf, logx.Lshortfile), "test")
	l.Notice("lineno")
	assert.Contains(t, buf.String(), "log_test.go:46")
}
