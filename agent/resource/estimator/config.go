package estimator

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
)

// Global estimator config
type GlobalConfig struct {
}

// Config
type Config struct {
	Ctx      context.Context
	Log      *logx.Log
	Provider *allocation.Provider
	Id       string // Full provider ID
}
