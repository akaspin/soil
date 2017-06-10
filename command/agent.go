package command

import (
	"context"
	"fmt"
	"github.com/akaspin/cut"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"github.com/akaspin/soil/agent/registry"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/agent/source"
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
	PoolSize   int
	Id         string
	Meta       []string
}

func (o *AgentOptions) Bind(cc *cobra.Command) {
	cc.Flags().StringArrayVarP(&o.ConfigPath, "config", "", []string{"/etc/soil/config.hcl"}, "configuration file")
	cc.Flags().IntVarP(&o.PoolSize, "pool", "", 4, "worker pool size")
	cc.Flags().StringVarP(&o.Id, "id", "", "localhost", "agent id")
	cc.Flags().StringArrayVarP(&o.Meta, "meta", "", nil, "node metadata")
}

type Agent struct {
	*cut.Environment
	*AgentOptions

	// reconfigurable
	config      *agent.Config
	privatePods []*manifest.Pod

	log             *logx.Log
	agentSource     *source.Map
	metaSource      *source.Map
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
	c.agentSource = source.NewMap(ctx, c.log, "agent", true, manifest.Constraint{
		"${agent.drain}": "false",
	})
	c.metaSource = source.NewMap(ctx, c.log, "meta", true, manifest.Constraint{})
	statusSource := source.NewStatus(ctx, c.log)
	sourceSv := supervisor.NewGroup(ctx,
		c.agentSource,
		c.metaSource,
		statusSource,
	)

	sink, schedulerSv := scheduler.New(
		ctx, c.log, c.PoolSize,
		[]agent.Source{c.agentSource, c.metaSource, statusSource},
		[]agent.AllocationReporter{statusSource},
	)
	c.privateRegistry = registry.NewPrivate(ctx, c.log, sink)

	// agent
	agentSV := supervisor.NewChain(ctx,
		sourceSv,
		schedulerSv,
		c.privateRegistry,
	)

	if err = agentSV.Open(); err != nil {
		return
	}

	c.configureArbiters()
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
				c.configureArbiters()
				c.configurePrivateRegistry()
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
		"drain":    fmt.Sprintf("%t", c.config.Drain),
		"pod_exec": c.config.Exec,
	}, true)
	c.metaSource.Set(c.config.Meta, true)
}

func (c *Agent) configurePrivateRegistry() (err error) {
	c.privateRegistry.Sync(c.privatePods)
	return
}
