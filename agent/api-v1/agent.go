package api_v1

import (
	"os"
	"github.com/akaspin/soil/api"
	"syscall"
)

func NewAgentStop(signalChan chan os.Signal) (w api.Endpoint) {
	w = NewWrapper(func() (err error) {
		defer func() {
			go func() {
				signalChan <- syscall.SIGTERM
			}()
		}()
		return
	})
	return
}
