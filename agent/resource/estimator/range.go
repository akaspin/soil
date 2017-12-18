package estimator

import (
	"context"
	"fmt"
	"github.com/RoaringBitmap/roaring"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/manifest"
	"strconv"
)

// Range estimator estimates integers from given range
type Range struct {
	ctx    context.Context
	cancel context.CancelFunc
	log    *logx.Log

	min uint32
	max uint32

	bitmap         *roaring.Bitmap
	allocations    map[string]rangeExecutorAllocation // allocation requests by id
	allocateChan   chan *allocation.Resource
	deallocateChan chan string

	downstream bus.Consumer
}

// Create new range estimator
func NewRange(ctx context.Context, log *logx.Log, provider *allocation.Provider, downstream bus.Consumer) (r *Range) {
	r = &Range{
		log:            log,
		bitmap:         roaring.New(),
		allocations:    map[string]rangeExecutorAllocation{},
		allocateChan:   make(chan *allocation.Resource),
		deallocateChan: make(chan string),
		downstream:     downstream,
	}
	r.ctx, r.cancel = context.WithCancel(ctx)
	if v, ok := provider.Config["min"]; ok {
		r.min = uint32(v.(int))
	}
	if v, ok := provider.Config["max"]; ok {
		r.max = uint32(v.(int))
	}

	go r.loop()
	return
}

func (r *Range) Close() error {
	r.cancel()
	return nil
}

// Allocate integer from range
func (r *Range) Allocate(request *allocation.Resource) (err error) {
	select {
	case <-r.ctx.Done():
		r.log.Warningf(`skip allocation request %v: %v`, request, r.ctx.Err())
		err = r.ctx.Err()
	case r.allocateChan <- request:
		r.log.Tracef(`allocation request accepted: %v`, request)
	}
	return
}

// Deallocate integer with given id
func (r *Range) Deallocate(name string) (err error) {
	select {
	case <-r.ctx.Done():
		r.log.Warningf(`skip deallocation request "%s": %v`, name, r.ctx.Err())
		err = r.ctx.Err()
	case r.deallocateChan <- name:
		r.log.Tracef(`deallocation request accepted: %v`, name)
	}
	return
}

func (r *Range) loop() {
	for {
		select {
		case <-r.ctx.Done():
			r.log.Debug("close")
			return
		case req := <-r.allocateChan:
			r.log.Tracef(`allocate req: %v`, req)
			r.allocate(req)
		case id := <-r.deallocateChan:
			r.log.Tracef(`deallocate req: %v`, id)
			r.deallocate(id)
		}
	}
}

func (r *Range) allocate(req *allocation.Resource) {
	id := req.Request.Name

	// try to find values in already allocated resources
	if allocated, ok := r.allocations[id]; ok && allocated.failure == nil {
		r.log.Tracef(`"id" is already allocated: %d`, allocated.value)
		return
	}

	// try to find recovered value
	var recoveredValue uint32

	if raw, ok := req.Values["value"]; ok {
		if parsed, parseErr := strconv.ParseUint(raw, 10, 32); parseErr != nil {
			r.log.Warningf(`can't parse value: %s:%s`, id, raw)
		} else {
			if recoveredValue = uint32(parsed); recoveredValue >= r.min && recoveredValue <= r.max {
				if ok = r.bitmap.CheckedAdd(recoveredValue); ok {
					r.log.Debugf(`"%s" allocated from recovery: %d`, id, recoveredValue)
					r.notify(id, rangeExecutorAllocation{
						value: recoveredValue,
					})
					return
				}
			} else {
				r.log.Warningf(`recovered value exceeds limits: %s: %d(min) < %d < %d(max)`, id, r.min, recoveredValue, r.max)
			}
		}
	}

	// not allocated or previous request is failed
	res, err := r.allocateBitmap()
	if err != nil {
		r.notify(id, rangeExecutorAllocation{
			failure: err,
		})
		r.log.Warningf(`fail to allocate "%s": %v`, id, err)
		return
	}
	r.notify(id, rangeExecutorAllocation{
		value: res,
	})
	r.log.Debugf(`allocated "%s": %d`, id, res)
}

func (r *Range) deallocate(id string) {
	if state, ok := r.allocations[id]; ok {
		if state.failure == nil {
			r.bitmap.Remove(state.value)
		}
		delete(r.allocations, id)
		r.log.Debugf(`deallocated: %s: %v`, id, state)
		r.downstream.ConsumeMessage(bus.NewMessage(id, nil))

		for allocatedId, alloc := range r.allocations {
			if alloc.failure != nil {
				res, err := r.allocateBitmap()
				if err != nil {
					r.notify(allocatedId, rangeExecutorAllocation{
						failure: err,
					})
					r.log.Warningf(`fail to allocate "%s": %v`, id, err)
					continue
				}
				r.notify(allocatedId, rangeExecutorAllocation{
					value: res,
				})
				r.log.Debugf(`allocated "%s": %d`, id, res)
			}
		}
		return
	}
	r.log.Tracef(`deallocate: not found: %s`, id)
}

func (r *Range) notify(id string, alloc rangeExecutorAllocation) {
	r.allocations[id] = alloc
	r.downstream.ConsumeMessage(NewEstimatorMessage(id, alloc.failure, manifest.Environment{
		"value": fmt.Sprintf("%d", alloc.value),
	}))
}

func (r *Range) allocateBitmap() (res uint32, err error) {
	if ok := r.bitmap.CheckedAdd(r.min); ok {
		res = r.min
		return
	}
	iter := r.bitmap.Iterator()
	for iter.HasNext() {
		candidate := iter.Next() + 1
		if candidate > r.max {
			err = NotAvailableError
			return
		}
		if ok := r.bitmap.CheckedAdd(candidate); ok {
			res = candidate
			return
		}
	}
	err = NotAvailableError
	return
}

type rangeExecutorAllocation struct {
	value   uint32
	failure error
}
