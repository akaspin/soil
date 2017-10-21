package agent

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/api"
	"github.com/akaspin/soil/agent/api/api-server"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/metrics"
	"github.com/akaspin/soil/agent/provision"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
)

type ServerOptions struct {
	AgentId    string
	ConfigPath []string
	Address    string
	Meta       map[string]string
}

// Agent instance
type Server struct {
	log     *logx.Log
	options ServerOptions

	sv                      supervisor.Component
	metaStorage             bus.Setter
	systemStorage           bus.Setter
	agentStorage            bus.Setter
	privateRegistryConsumer scheduler.RegistryConsumer
}

func NewServer(ctx context.Context, log *logx.Log, options ServerOptions) (s *Server) {
	s = &Server{
		log:     log.GetLog("server"),
		options: options,
	}

	var state allocation.Recovery
	if recoveryErr := state.FromFilesystem(allocation.DefaultSystemPaths(), allocation.DefaultDbusDiscoveryFunc); recoveryErr != nil {
		s.log.Errorf("recovered with failure: %v", recoveryErr)
	}

	systemPaths := allocation.DefaultSystemPaths()
	provisionArbiter := scheduler.NewArbiter(ctx, log, "provision", manifest.Constraint{
		"${agent.drain}": "!= true",
	})
	provisionDrainPipe := bus.NewDivertPipe(provisionArbiter, bus.NewMessage("0", map[string]string{"agent.drain": "true"}))
	provisionCompositePipe := bus.NewCompositePipe("0", provisionDrainPipe, "meta", "system")

	s.metaStorage = bus.NewStrictMapUpstream("meta", provisionCompositePipe)
	s.systemStorage = bus.NewStrictMapUpstream("system", provisionCompositePipe)
	s.agentStorage = bus.NewMapUpstream("agent", provisionCompositePipe)

	apiRouter := api_server.NewRouter(s.log,
		// status
		api.NewStatusPingGet(),
	)

	provisionEvaluator := provision.NewEvaluator(ctx, s.log, systemPaths, state, &metrics.BlackHole{})
	s.privateRegistryConsumer = scheduler.NewSink2(ctx, s.log, state,
		scheduler.NewBoundedEvaluator(provisionArbiter, provisionEvaluator))

	s.sv = supervisor.NewChain(ctx,
		supervisor.NewGroup(ctx, provisionEvaluator, provisionArbiter),
		s.privateRegistryConsumer.(supervisor.Component),
		api_server.NewServer(ctx, s.log, s.options.Address, apiRouter),
	)
	return
}

func (s *Server) Open() (err error) {
	if err = s.sv.Open(); err != nil {
		return
	}
	s.Configure()
	return
}

func (s *Server) Close() error {
	return s.sv.Close()
}

func (s *Server) Wait() (err error) {
	return s.sv.Wait()
}

func (s *Server) Configure() {
	s.log.Infof("config: %v", s.options)
	configReader, err := manifest.NewConfigReader(s.options.ConfigPath...)
	if err != nil {
		s.log.Errorf("error reading configs: %v", err)
	}
	serverCfg := DefaultConfig()
	serverCfg.Agent.Id = s.options.AgentId
	serverCfg.Meta = bus.CloneMap(s.options.Meta)
	if err = serverCfg.Unmarshal(configReader.GetReaders()...); err != nil {
		s.log.Errorf("error unmarshal configs: %v", err)
	}
	var registry manifest.Registry
	if err = registry.Unmarshal(manifest.PrivateNamespace, configReader.GetReaders()...); err != nil {
		s.log.Errorf("error unmarshal pods: %v", err)
	}
	s.metaStorage.Set(serverCfg.Meta)
	s.systemStorage.Set(serverCfg.System)
	s.privateRegistryConsumer.ConsumeRegistry(manifest.PrivateNamespace, registry)
	s.log.Debug("configure: done")
}
