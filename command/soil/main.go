package main

import (
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/command"
	"os"
)

func main() {
	err := command.Run(os.Stderr, os.Stdout, os.Stdin, os.Args[1:]...)
	if err != nil {
		logx.GetLog("main").Critical(err)
	}
}
