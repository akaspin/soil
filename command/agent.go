package command

import (
	"context"
	"github.com/akaspin/cut"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"github.com/akaspin/soil/agent/arbiter"
	"github.com/akaspin/soil/agent/registry"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/manifest"
	"github.com/akaspin/supervisor"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

type Agent struct {
	*cut.Environment
	*ConfigOptions

	// reconfigurable
	config *agent.Config
	privatePods []*manifest.Pod

	log *logx.Log
	agentArbiter *arbiter.MapArbiter
	metaArbiter *arbiter.MapArbiter
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

	// Arbiters (premature initialize)
	c.agentArbiter = arbiter.NewMapArbiter(ctx, c.log, "agent", true)
	c.metaArbiter = arbiter.NewMapArbiter(ctx, c.log, "meta", true)
	c.configureArbiters()

	sink, schedulerSV := scheduler.New(ctx, c.log, c.config.Workers, c.agentArbiter, c.metaArbiter)
	c.privateRegistry = registry.NewPrivate(ctx, c.log, sink)

	// agent
	agentSV := supervisor.NewChain(ctx,
		schedulerSV,
		c.privateRegistry,
	)

	if err = agentSV.Open(); err != nil {
		return
	}

	c.configurePrivateRegistry()

	// bind signals
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	LOOP:
	for {
		select {
		case sig := <-signalCh:
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				agentSV.Close()
				break LOOP
			case syscall.SIGHUP:
				c.log.Infof("SIGHUP received")
				c.readConfig()
				c.readPrivatePods()
				c.configurePrivateRegistry()
				c.configureArbiters()
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
	c.agentArbiter.Configure(map[string]string{
		"id": c.config.Id,
		"pod_exec": c.config.Exec,
	})
	c.metaArbiter.Configure(c.config.Meta)
}

func (c *Agent) configurePrivateRegistry() (err error) {
	c.privateRegistry.Sync(c.privatePods)
	return
}