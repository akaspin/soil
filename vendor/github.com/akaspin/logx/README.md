# logx 

[![GoDoc](https://godoc.org/github.com/akaspin/logx?status.svg)](http://godoc.org/github.com/akaspin/logx)
[![Build Status](https://img.shields.io/travis/akaspin/logx.svg)](https://travis-ci.org/akaspin/logx)
[![Codecov](https://img.shields.io/codecov/c/github/akaspin/logx/master.svg)](https://codecov.io/gh/akaspin/logx)
[![Go Report Card](https://goreportcard.com/badge/github.com/akaspin/logx)](https://goreportcard.com/report/github.com/akaspin/logx)

Logx is simple and fast logging library designed taking into account to work in the containerized environments. 

In the base Logx using go build tags to configure logging level.

## Usage

```go
package main

import "github.com/akaspin/logx"

func main() {
    log := logx.GetLog("test")
    
    log.Trace("too chatty")
    log.Debug("less chatty")
    log.Info("ok")
    log.Notice("something serious")
    log.Warning("oh")
    log.Error("bang")
}
```

By default all trace and debug calls are completely omitted from output:

```shell
$ go run ./main.go 
INFO test main.go:10 ok
NOTICE test main.go:11 something serious
WARNING test main.go:12 oh
ERROR test main.go:13 bang
```

To make app extremely chatty use "trace" build tag:

```shell
$ go run -tags=trace ./main.go 
TRACE test main.go:8 too chatty
DEBUG test main.go:9 less chatty
INFO test main.go:10 ok
NOTICE test main.go:11 something serious
WARNING test main.go:12 oh
ERROR test main.go:13 bang
```

In opposite to make app quiet use "notice":

```shell
$ go run -tags=notice ./main.go 
NOTICE test main.go:11 something serious
WARNING test main.go:12 oh
ERROR test main.go:13 bang
```



 

