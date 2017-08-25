package api

import (
	"context"
	"net/url"
)

// Endpoint handles RPC path-method
type Endpoint interface {

	// Empty returns empty structure or <nil> if endpoint doesn't receive data
	Empty() interface{}

	// Process handles URL and ingest structure and returns data or error
	Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error)
}

type Request interface {
}
