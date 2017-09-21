// +build trace

package logx_test

import (
	"bytes"
	"github.com/akaspin/logx"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStandaloneLogger_Trace_On(t *testing.T) {
	w := &bytes.Buffer{}
	l := logx.NewLog(logx.NewSimpleAppender(w, 0), "")
	l.Trace("test")
	assert.Equal(t, "TRACE test\n", w.String())
}
