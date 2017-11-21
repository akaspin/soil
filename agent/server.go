package agent

import (
	"context"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/api"
	"github.com/akaspin/soil/agent/api/api-server"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/cluster"
	"github.com/akaspin/soil/agent/provision"
	"github.com/akaspin/soil/agent/resource"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/lib"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/soil/proto"
	"github.com/akaspin/supervisor"
	"regexp"
)

var ServerVersion string

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

	sv supervisor.Component

	confPipe                bus.Consumer
	resourceEvaluator       *resource.Evaluator
	privateRegistryConsumer scheduler.RegistryConsumer
	kv                      *cluster.KV
}

func NewServer(ctx context.Context, log *logx.Log, options ServerOptions) (s *Server) {
	s = &Server{
		log:     log.GetLog("server"),
		options: options,
	}
	s.kv = cluster.NewKV(ctx, log, cluster.DefaultBackendFactory)

	var state allocation.Recovery
	if recoveryErr := state.FromFilesystem(allocation.DefaultSystemPaths(), allocation.DefaultDbusDiscoveryFunc); recoveryErr != nil {
		s.log.Errorf("recovered with failure: %v", recoveryErr)
	}

	systemPaths := allocation.DefaultSystemPaths()

	// Resource
	resourceArbiter := scheduler.NewArbiter(ctx, log, "resource", scheduler.ArbiterConfig{
		Required: manifest.Constraint{"${agent.drain}": "!= true"},
		ConstraintOnly: []*regexp.Regexp{
			regexp.MustCompile(`^evaluation\..+`),
		},
	})
	resourceDrainPipe := bus.NewDivertPipe(resourceArbiter, bus.NewMessage("private", map[string]string{"agent.drain": "true"}))
	resourceCompositePipe := bus.NewCompositePipe("private", log, resourceDrainPipe, "meta", "system", "resource")

	// provision
	provisionArbiter := scheduler.NewArbiter(ctx, log, "provision",
		scheduler.ArbiterConfig{
			Required: manifest.Constraint{"${agent.drain}": "!= true"},
			ConstraintOnly: []*regexp.Regexp{
				regexp.MustCompile(`^evaluation\..+`),
			},
		})
	provisionDrainPipe := bus.NewDivertPipe(provisionArbiter, bus.NewMessage("private", map[string]string{"agent.drain": "true"}))
	provisionCompositePipe := bus.NewCompositePipe("private", log, provisionDrainPipe, "meta", "system", "resource")

	s.confPipe = bus.NewTeePipe(resourceCompositePipe, provisionCompositePipe)

	drainFn := func(on bool) {
		resourceDrainPipe.Divert(on)
		provisionDrainPipe.Divert(on)
	}

	apiClusterNodesGet := api.NewClusterNodesGet(log)
	apiRouter := api_server.NewRouter(s.log,
		// status
		api.NewStatusPingGet(),

		// agent
		api.NewAgentReloadPut(s.Configure),
		api.NewAgentDrainPut(drainFn),
		api.NewAgentDrainDelete(drainFn),

		// cluster
		apiClusterNodesGet,
	)

	// watchers
	s.kv.Producer("nodes").Subscribe(ctx, bus.NewSlicerPipe(log, bus.NewTeePipe(
		apiRouter, apiClusterNodesGet.Processor().(bus.Consumer),
	)))

	s.resourceEvaluator = resource.NewEvaluator(ctx, log, resource.EvaluatorConfig{}, state, provisionCompositePipe, resourceCompositePipe)
	provisionEvaluator := provision.NewEvaluator(ctx, s.log, provision.EvaluatorConfig{
		SystemPaths: systemPaths,
		Recovery:    state,
	})
	s.privateRegistryConsumer = scheduler.NewSink(ctx, s.log, state,
		scheduler.NewBoundedEvaluator(resourceArbiter, s.resourceEvaluator),
		scheduler.NewBoundedEvaluator(provisionArbiter, provisionEvaluator),
	)

	s.sv = supervisor.NewChain(ctx,
		api_server.NewServer(ctx, s.log, s.options.Address, apiRouter),
		s.kv,
		supervisor.NewGroup(ctx, resourceArbiter, provisionArbiter),
		s.resourceEvaluator,
		provisionEvaluator,
		s.privateRegistryConsumer.(supervisor.Component),
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
	var buffers lib.StaticBuffers
	if err := buffers.ReadFiles(s.options.ConfigPath...); err != nil {
		s.log.Errorf("error reading configs: %v", err)

	}
	serverCfg := DefaultConfig()
	serverCfg.Meta = lib.CloneMap(s.options.Meta)

	if err := serverCfg.Unmarshal(buffers.GetReaders()...); err != nil {
		s.log.Errorf("unmarshal server configs: %v", err)
	}
	var resourceConfigs resource.Configs
	if err := resourceConfigs.Unmarshal(buffers.GetReaders()...); err != nil {
		s.log.Errorf("unmarshal resource configs: %v", err)
	}
	var registry manifest.Registry
	if err := registry.Unmarshal(manifest.PrivateNamespace, buffers.GetReaders()...); err != nil {
		s.log.Errorf("unmarshal registry: %v", err)
	}
	clusterConfig := cluster.DefaultConfig()
	clusterConfig.NodeID = s.options.AgentId
	if err := (&clusterConfig).Unmarshal(buffers.GetReaders()...); err != nil {
		s.log.Errorf("unmarshal cluster config: %v", err)
	}

	s.kv.Configure(clusterConfig)

	// announce node
	s.kv.VolatileStore("nodes").ConsumeMessage(bus.NewMessage("", proto.NodeInfo{
		ID:        clusterConfig.NodeID,
		Advertise: clusterConfig.Advertise,
		Version:   proto.Version,
		API:       proto.APIV1Version,
	}))

	s.confPipe.ConsumeMessage(bus.NewMessage("meta", serverCfg.Meta))
	s.confPipe.ConsumeMessage(bus.NewMessage("system", serverCfg.System))

	s.resourceEvaluator.Configure(resourceConfigs)
	s.privateRegistryConsumer.ConsumeRegistry(registry)
	s.log.Debug("configure: done")
}
