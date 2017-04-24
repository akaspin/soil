package concurrency

import "time"


// Config is common for all pools
type Config struct {
	Capacity int
	IdleTimeout time.Duration
	CloseTimeout time.Duration
}
