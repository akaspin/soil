package registry

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
)

type Private struct {
	*supervisor.Control
	log *logx.Log

	scheduler agent.Scheduler
}

func NewPrivate(ctx context.Context, log *logx.Log, scheduler agent.Scheduler) (p *Private) {
	p = &Private{
		Control:   supervisor.NewControl(ctx),
		log:       log.GetLog("registry", "private"),
		scheduler: scheduler,
	}
	return
}

func (r *Private) Sync(pods []*manifest.Pod) {
	r.scheduler.Sync("private", pods)
	return
}
