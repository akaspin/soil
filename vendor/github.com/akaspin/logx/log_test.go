package logx_test

import (
	"bytes"
	"github.com/akaspin/logx"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLog_Info(t *testing.T) {
	w := &bytes.Buffer{}
	l := logx.NewLog(logx.NewSimpleAppender(w, 0), "")
	l.Info("test")
	l.Infof("%s", "info format test")
	assert.Equal(t, "INFO test\nINFO info format test\n", w.String())
}

func TestLog_Prefix(t *testing.T) {
	w := &bytes.Buffer{}
	l := logx.NewLog(logx.NewSimpleAppender(w, logx.Lshortfile), "prefix")
	l.Warning("2")
	assert.Contains(t, w.String(), "WARNING prefix log_test.go")
	assert.Equal(t, "prefix", l.Prefix())
}

func TestLog_PrefixEmpty(t *testing.T) {
	w := &bytes.Buffer{}
	l := logx.NewLog(logx.NewSimpleAppender(w, logx.Lshortfile), "")
	l.Warning("2")
	assert.Contains(t, w.String(), "WARNING log_test.go")
	assert.Equal(t, "", l.Prefix())
}

func TestLog_GetLog(t *testing.T) {
	w := &bytes.Buffer{}
	l := logx.NewLog(logx.NewSimpleAppender(w, logx.Lshortfile), "")
	l.Warning("2")
	l2 := l.GetLog("second")
	l2.Info("test")
	assert.Contains(t, w.String(), "WARNING log_test.go:")
	assert.Contains(t, w.String(), "INFO second log_test.go:")
}
