package command

import (
	"context"
	"github.com/akaspin/cut"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"github.com/akaspin/soil/agent/api-v1"
	"github.com/akaspin/soil/agent/api-v1/api-server"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/bus/backend"
	"github.com/akaspin/soil/agent/bus/public"
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

	// bind signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// public KV
	publicBackend := backend.NewBackend(ctx, c.log, c.Public)
	publicRegistryPodsOperator := backend.NewPermanentOperator(publicBackend, "registry/pods")

	// public announcers
	publicNodeAnnouncer := backend.NewNodesAnnouncer(ctx, c.log, backend.NewEphemeralOperator(publicBackend, "nodes"), c.Id)

	manager := scheduler.NewManager(ctx, c.log,
		scheduler.NewManagerSource("agent", false, manifest.Constraint{
			"${agent.drain}": "!= true",
		}, "private", "public"),
		scheduler.NewManagerSource("system", false, nil, "private", "public"),
		scheduler.NewManagerSource("meta", false, nil, "private", "public"),
	)

	apiStatusNodeGet := api_v1.NewStatusNodeGet(c.log)
	apiStatusNodesGet := api_v1.NewStatusNodesGet(c.log)
	apiRegistryGet := api_v1.NewRegistryPodsGet(c.log)

	agentProducer := bus.NewFlatMap(ctx, c.log, false, "agent",
		manager,
		publicNodeAnnouncer,
		apiStatusNodeGet.Processor().(bus.MessageConsumer),
	)

	apiRouter := api_server.NewRouter(c.log,
		// status
		api_v1.NewStatusPingGet(),
		apiStatusNodesGet,
		apiStatusNodeGet,

		// lifecycle
		api_v1.NewAgentReloadPut(signalChan),
		api_v1.NewAgentStopPut(signalChan),
		api_v1.NewAgentDrainPut(agentProducer),
		api_v1.NewAgentDrainDelete(agentProducer),

		// registry
		apiRegistryGet,
		api_v1.NewRegistryPodsPut(c.log, publicRegistryPodsOperator),
		api_v1.NewRegistryPodsDelete(publicRegistryPodsOperator),
	)

	// private metadata
	metaProducer := bus.NewFlatMap(ctx, c.log, true, "meta",
		manager,
		apiStatusNodeGet.Processor().(bus.MessageConsumer),
	)
	systemProducer := bus.NewFlatMap(ctx, c.log, true, "system",
		manager,
		apiStatusNodeGet.Processor().(bus.MessageConsumer),
	)

	evaluator := scheduler.NewEvaluator(ctx, c.log)
	registrySink := scheduler.NewSink(ctx, c.log, evaluator, manager)

	// public watchers
	publicNodesWatcher := bus.NewPipe(ctx, c.log, "nodes", publicBackend, nil,
		apiStatusNodesGet.Processor().(bus.MessageConsumer),
		public.NewDiscoveryPipe(c.log, apiRouter),
	)
	publicRegistryWatcher := bus.NewPipe(ctx, c.log, "registry/pods", publicBackend, nil,
		public.NewRegistryWatcher(c.log,
			registrySink,
			apiRegistryGet.Processor().(bus.RegistryConsumer),
		))

	// SV
	agentSV := supervisor.NewChain(ctx,
		publicBackend,
		supervisor.NewGroup(ctx, evaluator, manager),
		supervisor.NewGroup(ctx, agentProducer, metaProducer, systemProducer),
		registrySink,
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
		registrySink,
		apiRegistryGet.Processor().(bus.RegistryConsumer),
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
				c.reload(systemProducer, metaProducer, registrySink,
					apiRegistryGet.Processor().(bus.RegistryConsumer))
			}
		case <-ctx.Done():
			break LOOP
		}
	}

	err = agentSV.Wait()
	c.log.Info("Bye")
	return
}

func (c *Agent) reload(systemSetter, metaSetter bus.Setter, registryConsumers ...bus.RegistryConsumer) (err error) {
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
	for _, registryConsumers := range registryConsumers {
		registryConsumers.ConsumeRegistry("private", private)
	}
	return
}
