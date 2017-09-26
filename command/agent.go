package command

import (
	"context"
	"github.com/akaspin/cut"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"github.com/akaspin/soil/agent/api-v1"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/bus/public"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/api"
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

	Public public.Options
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
	log             *logx.Log
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

	var agentProducer *bus.FlatMap

	// public KV
	publicBackend := public.NewBackend(ctx, c.log, c.Public)

	apiV1StatusNodes := api_v1.NewStatusNodes(c.log)
	apiV1StatusNode := api_v1.NewStatusNode(c.log)
	apiV1Registry := api_v1.NewRegistryGet(c.log)

	apiRouter := api.NewRouter(ctx, c.log,
		// status
		api.GET("/v1/status/ping", api_v1.NewPingEndpoint()),
		api.GET("/v1/status/nodes", apiV1StatusNodes),
		api.GET("/v1/status/node", apiV1StatusNode),

		// lifecycle
		api.GET("/v1/agent/reload", api_v1.NewWrapper(func() (err error) {
			signalChan <- syscall.SIGHUP
			return
		})),
		api.GET("/v1/agent/stop", api_v1.NewAgentStop(signalChan)),
		// drain
		api.PUT("/v1/agent/drain", api_v1.NewWrapper(func() (err error) {
			agentProducer.Set(map[string]string{
				"drain": "true",
			})
			return
		})),
		api.DELETE("/v1/agent/drain", api_v1.NewWrapper(func() (err error) {
			agentProducer.Set(map[string]string{
				"drain": "false",
			})
			return
		})),

		// registry
		api.GET("/v1/registry", apiV1Registry),
		api.PUT("/v1/registry", api_v1.NewRegistryPut(c.log, public.NewPermanentOperator(publicBackend, "registry"))),
	)


	// public announcers
	publicNodeAnnouncer := public.NewNodesAnnouncer(ctx, c.log, public.NewEphemeralOperator(publicBackend, "nodes"), c.Id)

	manager := scheduler.NewManager(ctx, c.log,
		scheduler.NewManagerSource("agent", false, manifest.Constraint{
			"${agent.drain}": "!= true",
		}, "private", "public"),
		scheduler.NewManagerSource("system", false, nil, "private", "public"),
		scheduler.NewManagerSource("meta", false, nil, "private", "public"),
	)

	// private metadata
	agentProducer = bus.NewFlatMap(ctx, c.log, false, "agent",
		manager,
		publicNodeAnnouncer,
		apiV1StatusNode,
	)
	metaProducer := bus.NewFlatMap(ctx, c.log, true, "meta",
		manager,
		apiV1StatusNode,
	)
	systemProducer := bus.NewFlatMap(ctx, c.log, true, "system",
		manager,
		apiV1StatusNode,
	)

	evaluator := scheduler.NewEvaluator(ctx, c.log)
	registrySink := scheduler.NewSink(ctx, c.log, evaluator, manager)

	// public watchers
	publicNodesWatcher := bus.NewPipe(ctx, c.log, "nodes", publicBackend, nil,
		apiV1StatusNodes,
		api.NewDiscoveryPipe(c.log, apiRouter),
	)
	publicRegistryWatcher := bus.NewPipe(ctx, c.log, "registry", publicBackend, nil,
		bus.NewWatcher(c.log, registrySink, apiV1Registry))

	// SV
	agentSV := supervisor.NewChain(ctx,
		publicBackend,
		supervisor.NewGroup(ctx, evaluator, manager),
		supervisor.NewGroup(ctx, agentProducer, metaProducer, systemProducer),
		registrySink,
		apiRouter,
		publicNodesWatcher,
		publicRegistryWatcher,
		api.NewServer(ctx, c.log, c.Address, apiRouter),
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

	c.reload(systemProducer, metaProducer, registrySink, apiV1Registry)

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
				c.reload(systemProducer, metaProducer, registrySink, apiV1Registry)
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
