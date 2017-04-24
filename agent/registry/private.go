package registry

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
)

type PrivateConfig struct {
	Pods []*manifest.Pod
}

type Private struct {
	*supervisor.Control
	log *logx.Log
	config PrivateConfig

	scheduler agent.Scheduler
}

func NewPrivate(ctx context.Context, log *logx.Log, scheduler agent.Scheduler, config PrivateConfig) (p *Private) {
	p = &Private{
		Control: supervisor.NewControl(ctx),
		log: log.GetLog("registry", "private"),
		config: config,
		scheduler: scheduler,
	}
	return
}

func (p *Private) Open() (err error) {
	p.scheduler.Sync(p.config.Pods)
	err = p.Control.Open()
	return
}

func (p *Private) Close() (err error) {
	err = p.Control.Close()
	return
}


func (p *Private) Wait() (err error) {
	err = p.Control.Wait()
	return
}
