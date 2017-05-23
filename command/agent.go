package command

import (
	"context"
	"github.com/akaspin/concurrency"
	"github.com/akaspin/cut"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"github.com/akaspin/soil/agent/arbiter"
	"github.com/akaspin/soil/agent/registry"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/agent/scheduler/executor"
	"github.com/akaspin/supervisor"
	"os"
	"os/signal"
	"syscall"
)

type Agent struct {
	*cut.Environment
	*ConfigOptions
}

func (c *Agent) Run(args ...string) (err error) {
	log := logx.GetLog("init")

	// parse configs
	config := agent.DefaultConfig()
	failures := config.Read(c.ConfigPath...)
	if len(failures) != 0 {
		log.Warningf("configuration parsed with errors %v", failures)
	}

	ctx := context.Background()

	// Executor
	pool := concurrency.NewWorkerPool(ctx, concurrency.Config{
		Capacity: config.Workers,
	})
	executorRt := executor.New(ctx, log, pool)
	executorSV := supervisor.NewChain(ctx, pool, executorRt)

	// Private

	privateFilter := arbiter.NewStatic(ctx, log, arbiter.StaticConfig{
		Id: config.Id,
		Meta: config.Meta,
		PodExec: config.Exec,
		Constraint: config.Local,
	})
	privateScheduler := scheduler.NewRuntime(ctx, log, executorRt, privateFilter, "private")
	privateRegistry := registry.NewPrivate(ctx, log, privateScheduler, registry.PrivateConfig{
		Pods: config.Local,
	})
	privateSV := supervisor.NewChain(ctx, privateFilter, privateScheduler, privateRegistry)

	// agent
	agentSV := supervisor.NewChain(ctx, executorSV, privateSV)

	if err = agentSV.Open(); err != nil {
		return
	}

	// bind signals
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-signalCh:
		agentSV.Close()
	case <-ctx.Done():
	}

	err = agentSV.Wait()
	log.Debug("exiting")
	return
}
