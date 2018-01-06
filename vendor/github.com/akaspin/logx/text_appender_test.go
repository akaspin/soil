package logx_test

import (
	"github.com/akaspin/logx"
	"strconv"
	"testing"
)

func BenchmarkBufferAppender_Append(b *testing.B) {
	b.Run(`simple`, func(b *testing.B) {
		output := &nopWriter{}
		a := logx.NewTextAppender(output, 0)
		for i := 0; i < b.N; i++ {
			a.Append("INFO", "test", strconv.Itoa(i))
		}
	})
	b.Run(`buf`, func(b *testing.B) {
		output := &nopWriter{}
		a := logx.NewTextAppender(output, 0)
		for i := 0; i < b.N; i++ {
			a.Append("INFO", "test", strconv.Itoa(i))
		}
	})
}
