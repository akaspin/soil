package logx_test

import (
	"bytes"
	"github.com/akaspin/logx"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStdSetOutput(t *testing.T) {
	w := &bytes.Buffer{}
	logx.SetAppender(logx.NewSimpleAppender(w, logx.Lshortfile))
	logx.Info("test")
	assert.Contains(t, w.String(), "INFO std_test.go")
}

func TestStdGetLogger(t *testing.T) {
	w := &bytes.Buffer{}
	logx.SetAppender(logx.NewSimpleAppender(w, logx.Lshortfile))
	logx.Info("test")

	l2 := logx.GetLog("second")
	l2.Info("test")
	assert.Contains(t, w.String(), "INFO std_test.go")
	assert.Contains(t, w.String(), "INFO second std_test.go")
}
