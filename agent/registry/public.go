package registry

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"github.com/akaspin/soil/agent/metadata"
	"github.com/akaspin/supervisor"
)

const publicRegistryPods = "registry/pod"

type Public struct {
	*supervisor.Control
	log *logx.Log

	producer  metadata.DynamicProducer
	scheduler agent.Scheduler
}

func NewPublic(ctx context.Context, log *logx.Log, producer metadata.DynamicProducer) (r *Public) {
	r = &Public{
		Control:  supervisor.NewControl(ctx),
		log:      log.GetLog("registry", "public"),
		producer: producer,
	}
	return
}

//func (r *Public) Open() (err error) {
//	r.producer.RegisterConsumer(publicRegistryPods, r)
//	err = r.Control.Open()
//	return
//}
