package command

import (
	"context"
	"github.com/akaspin/cut"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"github.com/akaspin/soil/agent/api-v1"
	"github.com/akaspin/soil/agent/public"
	"github.com/akaspin/soil/agent/registry"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/agent/metadata"
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

	Public public.BackendOptions
}

func (o *AgentOptions) Bind(cc *cobra.Command) {
	cc.Flags().StringVarP(&o.Id, "id", "", "localhost", "unique agent id")

	cc.Flags().StringArrayVarP(&o.ConfigPath, "config", "", []string{"/etc/soil/config.hcl"}, "configuration file")
	cc.Flags().StringArrayVarP(&o.Meta, "meta", "", nil, "node metadata in form field=value")
	cc.Flags().StringVarP(&o.Address, "address", "", ":7654", "listen address")

	cc.Flags().BoolVarP(&o.Public.Enabled, "public-enable", "", false, "enable public namespace clustering")
	cc.Flags().StringVarP(&o.Public.Advertise, "public-advertise", "", "127.0.0.1:7654", "advertise address public namespace")
	cc.Flags().StringVarP(&o.Public.URL, "public-backend", "", "consul://127.0.0.1:8500/soil", "backend url for public namespace")
	cc.Flags().DurationVarP(&o.Public.Timeout, "public-timeout", "", time.Second*30, "connect timeout for public namespace backend")
	cc.Flags().IntVarP(&o.Public.Retry, "public-retry", "", 0, "connection retry count for public namespace backend (0 to infinite)")
	cc.Flags().DurationVarP(&o.Public.RetryInterval, "public-retry-interval", "", time.Second*30, "public namespace backend connect retry interval")
	cc.Flags().DurationVarP(&o.Public.TTL, "public-ttl", "", time.Minute*5, "TTL for agent entries in public namespace backend")
}

type Agent struct {
	*cut.Environment
	*AgentOptions

	// reconfigurable
	config      *agent.Config
	privatePods []*manifest.Pod

	log             *logx.Log
	agentSource     *metadata.Plain
	metaSource      *metadata.Plain
	privateRegistry *registry.Private
}

func (c *Agent) Bind(cc *cobra.Command) {
	cc.Short = "Run agent"
}

func (c *Agent) Run(args ...string) (err error) {
	c.log = logx.GetLog("root")

	// parse configs
	c.readConfig()
	c.readPrivatePods()

	ctx := context.Background()

	// sources
	c.agentSource = metadata.NewPlain(ctx, c.log, "agent", false)
	c.metaSource = metadata.NewPlain(ctx, c.log, "meta", false)
	statusSource := metadata.NewAllocation(ctx, c.log)
	sourceSV := supervisor.NewGroup(ctx,
		c.agentSource,
		c.metaSource,
		statusSource,
	)

	//sink, manager, schedulerSv := scheduler.New(
	//	ctx, c.log,
	//	[]agent.EvaluationReporter{statusSource},
	//)
	manager := scheduler.NewManager(ctx, c.log)
	manager.AddProducer(c.agentSource, false, "private", "public")
	manager.AddProducer(c.metaSource, false, "private", "public")
	manager.AddProducer(statusSource, true, "private", "public")

	executor := scheduler.NewEvaluator(ctx, c.log)
	sink := scheduler.NewSink(ctx, c.log, executor, manager)

	schedulerSV := supervisor.NewChain(ctx,
		supervisor.NewGroup(ctx, executor, manager),
		sink,
	)

	c.privateRegistry = registry.NewPrivate(ctx, c.log, sink)

	// bind signals
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// Agent
	apiRouter := api.NewRouter()
	apiRouter.Get("/v1/agent/reload", api_v1.NewWrapper(func() (err error) {
		signalCh <- syscall.SIGHUP
		return
	}))
	apiRouter.Get("/v1/agent/stop", api_v1.NewWrapper(func() (err error) {
		signalCh <- syscall.SIGTERM
		return
	}))

	// Info
	statusInfoGetEndpoint := api_v1.NewStatusInfoGetEndpoint(ctx, c.agentSource, c.metaSource, statusSource)
	apiRouter.Get("/v1/status/ping", api_v1.NewPingEndpoint(c.Id))
	apiRouter.Get("/v1/status/info", statusInfoGetEndpoint)

	// drain
	apiRouter.Get("/v1/status/drain", api_v1.NewDrainGetEndpoint(c.Id, manager.DrainState))
	apiRouter.Put("/v1/agent/drain", api_v1.NewDrainPutEndpoint(manager.Drain))
	apiRouter.Delete("/v1/agent/drain", api_v1.NewDrainDeleteEndpoint(manager.Drain))

	apiServer := api.NewServer(ctx, c.log, c.Address, apiRouter)
	apiServerSV := supervisor.NewChain(ctx, statusInfoGetEndpoint, apiServer)

	// agent
	agentSV := supervisor.NewChain(ctx,
		sourceSV,
		schedulerSV,
		c.privateRegistry,
		apiServerSV,
	)

	if err = agentSV.Open(); err != nil {
		return
	}

	c.configureArbiters()
	c.configurePrivateRegistry()

LOOP:
	for {
		select {
		case sig := <-signalCh:
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				c.log.Infof("stop received")
				agentSV.Close()
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
	c.agentSource.Configure(map[string]string{
		"id":       c.Id,
		"pod_exec": c.config.Exec,
	})
	c.metaSource.Configure(c.config.Meta)
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
