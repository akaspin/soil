package api_server

import (
	"context"
	"net/url"
)

// Processor handles RPC path-method
type Processor interface {

	// Empty returns empty structure for request unmarshal or <nil> if request doesn't send any data.
	Empty() interface{}

	// Process handles URL and ingest structure and returns data or error
	Process(ctx context.Context, u *url.URL, v interface{}) (res interface{}, err error)
}
