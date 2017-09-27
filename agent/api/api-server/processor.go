package api_server

import (
	"context"
	"net/url"
)

// Processor handles RPC path-method
type Processor interface {

	// Empty returns empty structure or <nil> if processor doesn't receive data
	Empty() interface{}

	// Process handles URL and ingest structure and returns data or error
	Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error)
}
