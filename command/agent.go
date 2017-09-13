package command

import (
	"context"
	"github.com/akaspin/cut"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"github.com/akaspin/soil/agent/api-v1"
	"github.com/akaspin/soil/agent/registry"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/agent/source"
	"github.com/akaspin/soil/api"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type AgentOptions struct {
	ConfigPath []string
	Id         string
	Meta       []string
	Address    string
}

func (o *AgentOptions) Bind(cc *cobra.Command) {
	cc.Flags().StringArrayVarP(&o.ConfigPath, "config", "", []string{"/etc/soil/config.hcl"}, "configuration file")
	cc.Flags().StringVarP(&o.Id, "id", "", "localhost", "agent id")
	cc.Flags().StringArrayVarP(&o.Meta, "meta", "", nil, "node metadata")
	cc.Flags().StringVarP(&o.Address, "address", "", ":7654", "listen address")
}

type Agent struct {
	*cut.Environment
	*AgentOptions

	// reconfigurable
	config      *agent.Config
	privatePods []*manifest.Pod

	log             *logx.Log
	agentSource     *source.Plain
	metaSource      *source.Plain
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
	c.agentSource = source.NewPlain(ctx, c.log, "agent", true)
	c.metaSource = source.NewPlain(ctx, c.log, "meta", true)
	statusSource := source.NewStatus(ctx, c.log)
	sourceSv := supervisor.NewGroup(ctx,
		c.agentSource,
		c.metaSource,
		statusSource,
	)

	sink, arbiter, schedulerSv := scheduler.New(
		ctx, c.log,
		[]agent.Source{c.agentSource, c.metaSource, statusSource},
		[]agent.EvaluationReporter{statusSource},
	)
	c.privateRegistry = registry.NewPrivate(ctx, c.log, sink)

	// bind signals
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// API
	apiRouter := api.NewRouter()
	apiRouter.Get("/v1/agent/reload", api_v1.NewWrapper(func() (err error) {
		signalCh <- syscall.SIGHUP
		return
	}))
	apiRouter.Get("/v1/agent/stop", api_v1.NewWrapper(func() (err error) {
		signalCh <- syscall.SIGTERM
		return
	}))
	apiRouter.Get("/v1/status/ping", api_v1.NewWrapper(func() (err error) {
		return
	}))

	// drain
	apiRouter.Get("/v1/status/drain", api_v1.NewDrainGetEndpoint(c.Id, arbiter.DrainState))
	apiRouter.Put("/v1/agent/drain", api_v1.NewDrainPutEndpoint(arbiter.Drain))
	apiRouter.Delete("/v1/agent/drain", api_v1.NewDrainDeleteEndpoint(arbiter.Drain))

	apiServer := api.NewServer(ctx, c.log, c.Address, apiRouter)

	// agent
	agentSV := supervisor.NewChain(ctx,
		sourceSv,
		schedulerSv,
		c.privateRegistry,
		apiServer,
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
	c.agentSource.Set(map[string]string{
		"id":       c.Id,
		"pod_exec": c.config.Exec,
	}, true)
	c.metaSource.Set(c.config.Meta, true)
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
