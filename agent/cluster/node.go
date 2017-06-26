package cluster

import (
	"github.com/akaspin/logx"
	"github.com/akaspin/supervisor"
	"github.com/hashicorp/memberlist"
	"time"
)

// Node represents one node in cluster.
type Node struct {
	*supervisor.Control
	log *logx.Log

	// Permanent config
	Id               string
	BindAddr         string
	RpcAddr          string
	AdvertiseAddr    string
	AdvertiseRpcAddr string

	// Hot config
	join        []string
	joinTimeout time.Duration
	joinRetry   int

	ml *memberlist.Memberlist
}

// State represents node metadata and registry states
type State struct {
	Mark int64
	Timestamp int64
}