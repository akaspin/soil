package cluster

import (
	"github.com/mitchellh/hashstructure"
	"time"
)

// Backend config
type Config struct {
	// kind://address/chroot
	URL           string
	ID            string
	TTL           time.Duration
	RetryInterval time.Duration
}

func DefaultConfig() (c Config) {
	c = Config{
		URL:           "local://localhost/soil",
		ID:            "localhost",
		TTL:           time.Minute * 3,
		RetryInterval: time.Second * 30,
	}
	return
}

func (c Config) IsEqual(config Config) (res bool) {
	left, _ := hashstructure.Hash(c, nil)
	right, _ := hashstructure.Hash(config, nil)
	res = left == right
	return
}
