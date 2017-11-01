package cluster

import "time"

// Worker config
type Config struct {
	URL           string
	TTL           time.Duration
	RetryInterval time.Duration
}
