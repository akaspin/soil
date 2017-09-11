package registry

import (
	"github.com/akaspin/supervisor"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"context"
)

const publicRegistryPods  = "registry/pod"

type Public struct {
	*supervisor.Control
	log *logx.Log

	producer agent.SourceProducer
	scheduler agent.Scheduler
}

func NewPublic(ctx context.Context, log *logx.Log, producer agent.SourceProducer) (r *Public)  {
	r = &Public{
		Control: supervisor.NewControl(ctx),
		log: log.GetLog("registry", "public"),
		producer: producer,
	}
	return
}

func (r *Public) Open() (err error) {
	r.producer.RegisterConsumer(publicRegistryPods, r)
	err = r.Control.Open()
	return
}

func (r *Public) Sync(producer string, active bool, data map[string]string) {

}

