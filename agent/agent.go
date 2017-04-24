package agent

import (
	"github.com/akaspin/soil/manifest"
)

type Scheduler interface {
	Sync(pods []*manifest.Pod) (err error)
	Submit(name string, pod *manifest.Pod) (err error)
}

type Filter interface {
	// Register Pod manifest with given name and callback.
	Submit(name string, pod *manifest.Pod, fn func(reason error))

	// Get environment
	Environment() (res *Environment)
}