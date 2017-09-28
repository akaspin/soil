package resource

import (
	"github.com/RoaringBitmap/roaring"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"sync"
)

type Range struct {
	log      *logx.Log
	consumer bus.MessageConsumer

	allocationsMu sync.RWMutex
	allocations   map[string]*intRangeRequest

	poolMu sync.RWMutex
	min    int
	max    int
	pool   *roaring.Bitmap
}

func NewRange(log *logx.Log, consumer bus.MessageConsumer, id string) (r *Range) {
	r = &Range{
		log:         log.GetLog("resource", "range", id),
		consumer:    consumer,
		allocations: map[string]*intRangeRequest{},
	}
	return
}

// Submit accepts new request with id and given parameters
func (r *Range) Submit(id string, params map[string]interface{}) {
	panic("implement me")
}

type intRangeRequest struct {
	id        string
	isDynamic bool
	allocated int
}
