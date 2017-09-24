package registry

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
)

type Syncer interface {
	Sync(pods manifest.Pods)
}

type Private struct {
	*supervisor.Control
	log *logx.Log
	scheduler     agent.Scheduler
}

func New(ctx context.Context, log *logx.Log, scheduler agent.Scheduler) (p *Private) {
	p = &Private{
		Control:       supervisor.NewControl(ctx),
		log:           log.GetLog("registry", "private"),
		scheduler:     scheduler,
	}
	return
}

func (r *Private) Sync(pods []*manifest.Pod) {
	var verified []*manifest.Pod
	for _, pod := range pods {
		if err := pod.Verify("private"); err != nil {
			r.log.Error(err)
			continue
		}
		verified = append(verified, pod)
	}

	r.scheduler.Sync("private", verified)
	return
}
