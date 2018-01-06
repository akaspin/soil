package logx_test

import (
	"bytes"
	"strconv"
	"sync"
	"testing"
)

func BenchmarkParamAllocations(b *testing.B) {
	b.Run(`strings 64`, func(b *testing.B) {
		var a string
		fn := func(param string) {
			a = param
		}
		for i := 1; i < b.N; i++ {
			func() {
				p := []byte("string-longer-than-64-string-longer-than-64-string-longer-than-64")
				fn(string(p))
			}()
		}
	})
	b.Run(`bytes 64`, func(b *testing.B) {
		var a string
		fn := func(param []byte) {
			a = string(param)
		}
		for i := 1; i < b.N; i++ {
			func() {
				p := []byte("string-longer-than-64-string-longer-than-64-string-longer-than-64")
				fn(p)
			}()
		}
	})
	b.Run(`strings 16`, func(b *testing.B) {
		var a string
		fn := func(param string) {
			a = param
		}
		for i := 1; i < b.N; i++ {
			func() {
				p := []byte("string-longer-than-16")
				fn(string(p))
			}()
		}
	})
	b.Run(`bytes 16`, func(b *testing.B) {
		var a string
		fn := func(param []byte) {
			a = string(param)
		}
		for i := 1; i < b.N; i++ {
			func() {
				p := []byte("string-longer-than-16")
				fn(p)
			}()
		}
	})
}

type nopWriter struct{}

func (*nopWriter) Write(p []byte) (n int, err error) {
	return
}

func BenchmarkBufferPool(b *testing.B) {
	b.Run(`byte slice`, func(b *testing.B) {
		var res nopWriter
		for i := 0; i < b.N; i++ {
			func() {
				var chunk []byte
				for j := 0; j < 128; j++ {
					chunk = append(chunk, []byte(strconv.Itoa(j))...)
				}
				res.Write(chunk)
			}()
		}
	})
	b.Run(`buffer pool`, func(b *testing.B) {
		var res nopWriter
		pool := &sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		}
		for i := 0; i < b.N; i++ {
			func() {
				chunk := pool.Get().(*bytes.Buffer)
				for j := 0; j < 128; j++ {
					chunk.Write([]byte(strconv.Itoa(j)))
				}
				chunk.WriteTo(&res)
				chunk.Reset()
				pool.Put(chunk)
			}()
		}
	})
}
