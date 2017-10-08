package command

import (
	"context"
	"github.com/akaspin/cut"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/api"
	"github.com/akaspin/soil/agent/api/api-server"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/bus/backend"
	"github.com/akaspin/soil/agent/bus/public"
	"github.com/akaspin/soil/agent/metrics"
	"github.com/akaspin/soil/agent/provision"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type AgentOptions struct {
	Id string // Unique Agent ID

	ConfigPath []string // Paths to configuration files
	Meta       []string // Metadata set
	Address    string   // bind address

	Public backend.Options
}

func (o *AgentOptions) Bind(cc *cobra.Command) {
	cc.Flags().StringVarP(&o.Id, "id", "", "localhost", "unique agent id")

	cc.Flags().StringArrayVarP(&o.ConfigPath, "config", "", []string{"/etc/soil/config.hcl"}, "configuration file")
	cc.Flags().StringArrayVarP(&o.Meta, "meta", "", nil, "node metadata in form field=value")
	cc.Flags().StringVarP(&o.Address, "address", "", ":7654", "listen address")

	cc.Flags().BoolVarP(&o.Public.Enabled, "public", "", false, "enable public namespace clustering")
	cc.Flags().StringVarP(&o.Public.Advertise, "advertise", "", "127.0.0.1:7654", "advertise address public namespace")
	cc.Flags().StringVarP(&o.Public.URL, "url", "", "consul://127.0.0.1:8500/soil", "url for public backend")
	cc.Flags().DurationVarP(&o.Public.TTL, "ttl", "", time.Minute*3, "TTL for dynamic entries in public backend")

	cc.Flags().DurationVarP(&o.Public.Timeout, "timeout", "", time.Minute, "connect timeout for public backend")
	cc.Flags().DurationVarP(&o.Public.RetryInterval, "interval", "", time.Second*30, "public backend connect retry interval")
}

type Agent struct {
	*cut.Environment
	*AgentOptions
	log *logx.Log
}

func (c *Agent) Bind(cc *cobra.Command) {
	cc.Short = "Run agent"
}

func (c *Agent) Run(args ...string) (err error) {
	ctx := context.Background()
	c.log = logx.GetLog("root")

	systemPaths := allocation.DefaultSystemPaths()
	var state allocation.Recovery
	if recoveryErr := state.FromFilesystem(allocation.DefaultSystemPaths(), allocation.DefaultDbusDiscoveryFunc); recoveryErr != nil {
		c.log.Errorf("recovered with failure: %v", recoveryErr)
	}

	// bind signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// public KV
	publicBackend := backend.NewLibKVBackend(ctx, c.log, c.Public)
	publicRegistryPodsOperator := backend.NewPermanentOperator(publicBackend, "registry/pods")

	// public announcers
	publicNodeAnnouncer := public.NewNodeAnnouncer(c.log, backend.NewEphemeralOperator(publicBackend, "nodes"), c.Id)

	provisionManager := scheduler.NewManager(ctx, c.log, "provision",
		scheduler.NewManagerSource("agent", false, manifest.Constraint{
			"${agent.drain}": "!= true",
		}, "private", "public"),
		scheduler.NewManagerSource("system", false, nil, "private", "public"),
		scheduler.NewManagerSource("meta", false, nil, "private", "public"),
	)

	apiStatusNodeGet := api.NewStatusNodeGet(c.log)
	apiStatusNodesGet := api.NewStatusNodesGet(c.log)
	apiRegistryGet := api.NewRegistryPodsGet(c.log)

	agentProducer := bus.NewMapUpstream("agent",
		provisionManager,
		publicNodeAnnouncer,
		apiStatusNodeGet.Processor().(bus.MessageConsumer),
	)
	// private metadata
	metaProducer := bus.NewStrictMapUpstream("meta",
		provisionManager,
		apiStatusNodeGet.Processor().(bus.MessageConsumer),
	)
	systemProducer := bus.NewStrictMapUpstream("system",
		provisionManager,
		apiStatusNodeGet.Processor().(bus.MessageConsumer),
	)

	apiRouter := api_server.NewRouter(c.log,
		// status
		api.NewStatusPingGet(),
		apiStatusNodesGet,
		apiStatusNodeGet,

		// lifecycle
		api.NewAgentReloadPut(signalChan),
		api.NewAgentStopPut(signalChan),
		api.NewAgentDrainPut(agentProducer),
		api.NewAgentDrainDelete(agentProducer),

		// registry
		apiRegistryGet,
		api.NewRegistryPodsPut(c.log, publicRegistryPodsOperator),
		api.NewRegistryPodsDelete(publicRegistryPodsOperator),
	)

	provisionEvaluator := provision.NewEvaluator(ctx, c.log, systemPaths, state, &metrics.BlackHole{})
	sink := scheduler.NewSink(ctx, c.log, state,
		scheduler.NewManagedEvaluator(provisionManager, provisionEvaluator))

	// public watchers
	publicNodesWatcher := bus.NewBoundedPipe(ctx, c.log, "nodes", publicBackend, nil,
		apiStatusNodesGet.Processor().(bus.MessageConsumer),
		public.NewDiscoveryPipe(c.log, apiRouter),
	)
	publicRegistryWatcher := bus.NewBoundedPipe(ctx, c.log, "registry/pods", publicBackend, nil,
		public.NewRegistryWatcher(c.log,
			sink,
			apiRegistryGet.Processor().(scheduler.RegistryConsumer),
		))

	// SV
	agentSV := supervisor.NewChain(ctx,
		publicBackend,
		supervisor.NewGroup(ctx, provisionEvaluator, provisionManager),
		sink,
		publicNodesWatcher,
		publicRegistryWatcher,
		api_server.NewServer(ctx, c.log, c.Address, apiRouter),
	)

	if err = agentSV.Open(); err != nil {
		return
	}

	// Configure static agent properties
	agentProducer.Set(map[string]string{
		"id":        c.Id,
		"advertise": c.Public.Advertise,
		"drain":     "false",
		"version":   V,
		"api":       "v1",
	})

	c.reload(systemProducer, metaProducer,
		sink,
		apiRegistryGet.Processor().(scheduler.RegistryConsumer),
	)

LOOP:
	for {
		select {
		case sig := <-signalChan:
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				c.log.Infof("stop received")
				go agentSV.Close()
				break LOOP
			case syscall.SIGHUP:
				c.log.Infof("reload received")
				c.reload(systemProducer, metaProducer, sink,
					apiRegistryGet.Processor().(scheduler.RegistryConsumer))
			}
		case <-ctx.Done():
			break LOOP
		}
	}

	err = agentSV.Wait()
	c.log.Info("Bye")
	return
}

func (c *Agent) reload(systemSetter, metaSetter bus.Setter, registryConsumers ...scheduler.RegistryConsumer) {
	// read config
	config := agent.DefaultConfig()
	for _, meta := range c.Meta {
		split := strings.SplitN(meta, "=", 2)
		if len(split) != 2 {
			c.log.Warningf("bad --meta=%s", meta)
			continue
		}
		config.Meta[split[0]] = split[1]
	}
	if err := config.Read(c.ConfigPath...); err != nil {
		c.log.Warningf("error reading config %s", err)
	}

	// producers
	systemSetter.Set(config.System)
	metaSetter.Set(config.Meta)

	// private registry
	var private manifest.Registry
	if err := private.UnmarshalFiles("private", c.ConfigPath...); err != nil {
		c.log.Warningf("error reading private registry %s", err)
	}
	for _, registryConsumer := range registryConsumers {
		registryConsumer.ConsumeRegistry("private", private)
	}
	return
}
