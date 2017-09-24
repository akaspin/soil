package command

import (
	"context"
	"fmt"
	"github.com/akaspin/cut"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"github.com/akaspin/soil/agent/api-v1"
	"github.com/akaspin/soil/agent/metadata"
	"github.com/akaspin/soil/agent/public"
	"github.com/akaspin/soil/agent/public/kv"
	"github.com/akaspin/soil/agent/registry"
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

	Public kv.Options
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

	// reconfigurable
	config      *agent.Config
	privatePods []*manifest.Pod

	log             *logx.Log
	agentProducer   *metadata.SimpleProducer
	metaProducer    *metadata.SimpleProducer
	privateRegistry *registry.Private
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

	// parse configs
	c.readConfig()
	c.readPrivatePods()

	/// API

	// public KV
	publicBackend := kv.NewBackend(ctx, c.log, c.Public)

	apiV1StatusNodes := api_v1.NewStatusNodes(c.log)
	apiV1StatusNode := api_v1.NewStatusNode(c.log)

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
			c.agentProducer.Set(true, map[string]string{
				"drain": "true",
			})
			return
		})),
		api.DELETE("/v1/agent/drain", api_v1.NewWrapper(func() (err error) {
			c.agentProducer.Set(true, map[string]string{
				"drain": "false",
			})
			return
		})),

		// registry
		api.PUT("/v1/registry", api_v1.NewRegistryPut(c.log, publicBackend)),
	)

	// public watchers
	publicNodesWatcher := metadata.NewPipe(ctx, c.log, "nodes", publicBackend, nil,
		apiV1StatusNodes,
		api.NewDiscoveryPipe(c.log, apiRouter),
	)
	publicRegistryWatcher := metadata.NewPipe(ctx, c.log, "registry", publicBackend, nil)

	// public announcers
	publicNodeAnnouncer := public.NewNodesAnnouncer(ctx, c.log, publicBackend, fmt.Sprintf("nodes/%s", c.Id))

	manager := metadata.NewManager(ctx, c.log,
		metadata.NewManagerSource("agent", false, manifest.Constraint{
			"${agent.drain}": "!= true",
		}, "private", "public"),
		metadata.NewManagerSource("meta", false, nil, "private", "public"),
		metadata.NewManagerSource("private_registry", true, nil, "private"),
		metadata.NewManagerSource("registry.public.valid", true, nil, "public"),
	)

	// private metadata
	c.agentProducer = metadata.NewSimpleProducer(ctx, c.log, "agent",
		manager,
		publicNodeAnnouncer,
		apiV1StatusNode,
	)
	c.metaProducer = metadata.NewSimpleProducer(ctx, c.log, "meta",
		manager,
		apiV1StatusNode,
	)

	evaluator := scheduler.NewEvaluator(ctx, c.log)
	registrySink := scheduler.NewSink(ctx, c.log, evaluator, manager)
	c.privateRegistry = registry.New(ctx, c.log, registrySink, manager)

	// SV
	agentSV := supervisor.NewChain(ctx,
		publicBackend,
		supervisor.NewGroup(ctx, evaluator, manager),
		supervisor.NewGroup(ctx, c.agentProducer, c.metaProducer),
		registrySink,
		c.privateRegistry,
		apiRouter,
		publicNodesWatcher,
		publicRegistryWatcher,
		api.NewServer(ctx, c.log, c.Address, apiRouter),
	)

	if err = agentSV.Open(); err != nil {
		return
	}

	c.agentProducer.Replace(map[string]string{
		"id":        c.Id,
		"advertise": c.Public.Advertise,
		"pod_exec":  c.config.Exec,
		"drain":     "false",
		"version":   V,
		"api":       "v1",
	})
	c.configureArbiters()
	c.configurePrivateRegistry()

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
				c.reload()
			}
		case <-ctx.Done():
			break LOOP
		}
	}

	err = agentSV.Wait()
	c.log.Info("Bye")
	return
}

func (c *Agent) readConfig() {
	c.config = agent.DefaultConfig()
	for _, meta := range c.Meta {
		split := strings.SplitN(meta, "=", 2)
		if len(split) != 2 {
			c.log.Warningf("bad --meta=%s", meta)
			continue
		}
		c.config.Meta[split[0]] = split[1]
	}
	if err := c.config.Read(c.ConfigPath...); err != nil {
		c.log.Warningf("error reading config %s", err)
	}
	return
}

func (c *Agent) readPrivatePods() {
	var err error
	if c.privatePods, err = manifest.ParseFromFiles("private", c.ConfigPath...); err != nil {
		c.log.Warningf("error reading private pods %s", err)
	}
}

func (c *Agent) configureArbiters() {
	c.metaProducer.Replace(c.config.Meta)
}

func (c *Agent) configurePrivateRegistry() (err error) {
	c.privateRegistry.Sync(c.privatePods)
	return
}

func (c *Agent) reload() (err error) {
	c.readConfig()
	c.readPrivatePods()
	c.configureArbiters()
	c.configurePrivateRegistry()
	return
}
