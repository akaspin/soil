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

//func TestLog_SetOutput(t *testing.T) {
//	w1 := &bytes.Buffer{}
//	l := logx.NewLog(logx.NewSimpleAppender(w1, logx.Lshortfile), "")
//	l.Info("1")
//
//	w2 := &bytes.Buffer{}
//	l.SetOutput(w2)
//	l.Warning("2")
//
//	assert.Contains(t, w1.String(), "INFO log_test.go")
//	assert.NotContains(t, w1.String(), "WARNING log_test.go")
//
//	assert.Contains(t, w2.String(), "WARNING log_test.go")
//	assert.NotContains(t, w2.String(), "INFO log_test.go")
//}

func TestLog_GetLog(t *testing.T) {
	w := &bytes.Buffer{}
	l := logx.NewLog(logx.NewSimpleAppender(w, logx.Lshortfile), "")
	l.Warning("2")
	l2 := l.GetLog("second")
	l2.Info("test")
	assert.Contains(t, w.String(), "WARNING log_test.go:")
	assert.Contains(t, w.String(), "INFO second log_test.go:")
}
