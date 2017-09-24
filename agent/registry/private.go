package registry

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"github.com/akaspin/soil/agent/metadata"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
)

type Syncer interface {
	Sync(pods manifest.Pods)
}

type Private struct {
	*supervisor.Control
	log *logx.Log

	stateConsumer metadata.Consumer
	scheduler     agent.Scheduler
}

func New(ctx context.Context, log *logx.Log, scheduler agent.Scheduler, stateConsumer metadata.Consumer) (p *Private) {
	p = &Private{
		Control:       supervisor.NewControl(ctx),
		log:           log.GetLog("registry", "private"),
		stateConsumer: stateConsumer,
		scheduler:     scheduler,
	}
	return
}

func (r *Private) Sync(pods []*manifest.Pod) {
	r.stateConsumer.ConsumeMessage(metadata.Message{
		Prefix: "private_registry",
		Clean:  false,
	})
	defer r.stateConsumer.ConsumeMessage(metadata.Message{
		Prefix: "private_registry",
		Clean:  true,
	})
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
