package command

import (
	"context"
	"github.com/akaspin/cut"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type AgentOptions struct {
	ServerOptions agent.ServerOptions
	Meta          []string // Metadata set
}

func (o *AgentOptions) Bind(cc *cobra.Command) {
	cc.Flags().StringVarP(&o.ServerOptions.AgentId, "id", "", "", "agent id (deprecated)")
	cc.Flags().StringArrayVarP(&o.ServerOptions.ConfigPath, "config", "", []string{"/etc/soil/config.hcl"}, "configuration file")
	cc.Flags().StringArrayVarP(&o.Meta, "meta", "", nil, "node metadata in form field=value")
	cc.Flags().StringVarP(&o.ServerOptions.Address, "address", "", ":7654", "listen address")
}

type Agent struct {
	*cut.Environment
	*AgentOptions
}

func (a *Agent) Bind(cc *cobra.Command) {
	cc.Short = "Run agent"
}

func (a *Agent) Run(args ...string) (err error) {
	ctx := context.Background()
	log := logx.GetLog("root")

	for _, metaChunk := range a.Meta {
		split := strings.SplitN(metaChunk, "=", 2)
		if len(split) != 2 {
			log.Warningf("bad --meta=%s", metaChunk)
			continue
		}
		a.ServerOptions.Meta[split[0]] = split[1]
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	server := agent.NewServer(ctx, log, a.ServerOptions)
	if err = server.Open(); err != nil {
		return
	}
	server.Configure()

LOOP:
	for {
		select {
		case sig := <-signalChan:
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				log.Infof("stop received")
				go server.Close()
				break LOOP
			case syscall.SIGHUP:
				log.Infof("reload received")
				server.Configure()
			}
		case <-ctx.Done():
			break LOOP
		}
	}

	err = server.Wait()
	log.Info("Bye")

	return
}
