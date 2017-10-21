package resource

import (
	"fmt"
	"github.com/RoaringBitmap/roaring"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"strconv"
	"sync"
)

var (
	executorNotAvailableError = fmt.Errorf("not-available")
)

type RangeExecutor struct {
	log      *logx.Log
	consumer bus.MessageConsumer
	min      uint32
	max      uint32

	mu          sync.Mutex
	state       *roaring.Bitmap
	allocations map[string]rangeExecutorAllocation
}

func NewRangeExecutor(log *logx.Log, config ExecutorConfig, consumer bus.MessageConsumer) (e *RangeExecutor, err error) {
	e = &RangeExecutor{
		log:         log,
		consumer:    consumer,
		max:         ^uint32(0),
		state:       roaring.New(),
		allocations: map[string]rangeExecutorAllocation{},
	}
	if v, ok := config.Properties["min"]; ok {
		e.min = uint32(v.(int))
	}
	if v, ok := config.Properties["max"]; ok {
		e.max = uint32(v.(int))
	}
	e.log.Debugf("started: min:%d max:%d", e.min, e.max)
	return
}

func (e *RangeExecutor) Close() error {
	return nil
}

func (e *RangeExecutor) Allocate(request Alloc) {
	e.log.Tracef(`request: %v`, request)
	var value uint32
	var recovered bool
	id := request.GetID()

	// try to recover value from allocated
	if val, ok := request.Values.GetPayload()["value"]; ok {
		if parsed, err := strconv.ParseUint(val, 10, 32); err != nil {
			e.log.Warningf(`can't parse value: %s:%s`, id, val)
		} else {
			if value = uint32(parsed); value >= e.min && value <= e.max {
				recovered = true
				e.log.Tracef(`recovered value: %s:%d`, id, value)
			} else {
				e.log.Warningf(`recovered value exceeds limits: %s:%d:%d:%d`, id, e.min, value, e.max)
			}
		}
	}
	go func(id string, value uint32, recovered bool) {
		e.mu.Lock()
		defer e.mu.Unlock()
		// alloc found
		if state, ok := e.allocations[id]; ok {
			e.log.Tracef(`found: %s:%v`, id, state)
			if state.failure == nil {
				e.log.Tracef(`already allocated: %s:%v`, id, state)
				return
			}
		}
		// recovered
		if recovered {
			if ok := e.state.CheckedAdd(value); ok {
				e.notify(id, rangeExecutorAllocation{value: value})
				return
			}
		}
		e.allocateValue(id)
	}(id, value, recovered)
}

func (e *RangeExecutor) Deallocate(id string) {
	go func(id string) {
		e.mu.Lock()
		defer e.mu.Unlock()
		if state, ok := e.allocations[id]; ok {
			if state.failure == nil {
				e.state.Remove(state.value)
			}
			delete(e.allocations, id)
			e.log.Debugf(`deallocated: %s:%v`, id, state)
			e.consumer.ConsumeMessage(bus.NewMessage(id, nil))
			for allocatedId, allocation := range e.allocations {
				if allocation.failure != nil {
					e.allocateValue(allocatedId)
				}
			}
			return
		}
		e.log.Tracef(`deallocate: not found: %s`, id)
	}(id)
}

func (e *RangeExecutor) allocateValue(id string) {
	if ok := e.state.CheckedAdd(e.min); ok {
		e.notify(id, rangeExecutorAllocation{
			value: e.min,
		})
		return
	}
	iter := e.state.Iterator()
	for iter.HasNext() {
		candidate := iter.Next() + 1
		if candidate > e.max {
			e.notify(id, rangeExecutorAllocation{
				failure: executorNotAvailableError,
			})
			return
		}
		if ok := e.state.CheckedAdd(candidate); ok {
			e.notify(id, rangeExecutorAllocation{
				value: candidate,
			})
			return
		}
	}
}

func (e *RangeExecutor) notify(id string, state rangeExecutorAllocation) {
	e.allocations[id] = state
	e.log.Debugf(`allocated %s:%d`, id, state.value)
	e.consumer.ConsumeMessage(NewExecutorMessage(id, state.failure, map[string]string{
		"value": fmt.Sprintf("%d", state.value),
	}))
}

type rangeExecutorAllocation struct {
	value   uint32
	failure error
}
