package registry

import (
	"github.com/akaspin/supervisor"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"context"
	"github.com/akaspin/soil/agent/metadata"
)

const publicRegistryPods  = "registry/pod"

type Public struct {
	*supervisor.Control
	log *logx.Log

	producer  metadata.Producer
	scheduler agent.Scheduler
}

func NewPublic(ctx context.Context, log *logx.Log, producer metadata.Producer) (r *Public)  {
	r = &Public{
		Control: supervisor.NewControl(ctx),
		log: log.GetLog("registry", "public"),
		producer: producer,
	}
	return
}

//func (r *Public) Open() (err error) {
//	r.producer.RegisterConsumer(publicRegistryPods, r)
//	err = r.Control.Open()
//	return
//}


