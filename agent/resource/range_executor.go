package resource

import (
	"github.com/RoaringBitmap/roaring"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"sync"
	"strconv"
	"fmt"
)

var (
	executorOutOfRangeError = fmt.Errorf("out-of-range")
	executorNotAvailableError = fmt.Errorf("not-available")
)

type RangeExecutor struct {
	log *logx.Log
	consumer bus.MessageConsumer
	min uint32
	max uint32

	mu sync.Mutex
	state *roaring.Bitmap
	allocations map[string]rangeExecutorAllocation
}

func NewRangeExecutor(log *logx.Log, config ExecutorConfig, consumer bus.MessageConsumer) (e *RangeExecutor, err error)  {
	e = &RangeExecutor{
		log: log,
		consumer: consumer,
		max: ^uint32(0),
		state: roaring.New(),
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
	var err error
	var value uint32
	var fixed bool
	id := request.GetID()
	// try to recover value from allocated
	if val, ok := request.Values.GetPayload()["value"]; ok {
		var parsed uint64
		parsed, err = strconv.ParseUint(val, 10, 32)
		if err != nil {
			e.log.Warningf(`[%s] can't parse value: %s`, id, val)
		} else {
			value = uint32(parsed)
			e.log.Tracef(`[%s] recovered value: %d`, id, value)
		}
	}
	// try to parse config
	if val, ok := request.Request.Config["value"]; ok {
		parsed := uint32(val.(int))
		fixed = true
		if value != parsed {
			e.log.Tracef(`[%s] not equal: %d(config) != %d(value)`, id, parsed, value)
			value = parsed
		}
		e.log.Tracef(`[%s] use fixed value: %d`, id, value)
	}
	go e.allocate(id, value, fixed)
}

func (e *RangeExecutor) Deallocate(id string) {
	e.log.Tracef(`[%s] deallocate`, id)
	e.mu.Lock()
	defer e.mu.Unlock()
	if state, ok := e.allocations[id]; ok {
		if state.failure == nil {
			e.state.Remove(state.value)
		}
		delete(e.allocations, id)
		e.consumer.ConsumeMessage(bus.NewMessage(id, nil))
		for fid, falloc := range e.allocations {
			if falloc.failure == executorNotAvailableError {
				if falloc.fixed {
					e.allocFixed(fid, falloc.value)
				} else {
					e.allocDynamic(fid)
				}
			}
		}
		return
	}
	e.log.Tracef(`[%s] deallocate: not found`, id)
}

func (e *RangeExecutor) allocate(id string, value uint32, fixed bool) {
	e.log.Tracef(`[%s] allocate: %d %t`, id, value, fixed)
	e.mu.Lock()
	defer e.mu.Unlock()
	var state rangeExecutorAllocation
	var ok bool

	if state, ok = e.allocations[id]; ok {
		e.log.Tracef(`[%s] allocation found: %v`, id, state)
		if state.failure == nil && (!fixed || fixed && state.value == value) {
			e.log.Tracef(`[%s] already allocated: %d`, id, value)
			return
		}
		if fixed && state.failure == nil && state.value != value {
			// migrate
			e.log.Tracef(`[%s] migrate fixed: %d->%d`, id, state.value, value)
			e.state.Remove(state.value)
			e.allocFixed(id, value)
			return
		}
		if fixed && state.failure != nil {
			e.log.Tracef(`[%s] try allocate failed (%v): %d`, id, state.failure, value)
			e.allocFixed(id, value)
			return
		}
		return
	}
	if fixed {
		e.allocFixed(id, value)
	} else {
		e.allocDynamic(id)
	}
}

func (e *RangeExecutor) allocFixed(id string, value uint32) {
	if value < e.min || value > e.max {
		e.notify(id, rangeExecutorAllocation{
			failure: executorOutOfRangeError,
		})
		return
	}
	if ok := e.state.CheckedAdd(value); ok {
		e.notify(id, rangeExecutorAllocation{
			value: value,
		})
		return
	}
	e.notify(id, rangeExecutorAllocation{
		value: value,
		fixed: true,
		failure: executorNotAvailableError,
	})
}

func (e *RangeExecutor) allocDynamic(id string) {
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
	e.consumer.ConsumeMessage(NewExecutorMessage(id, state.failure, map[string]string{
		"value": fmt.Sprintf("%d", state.value),
	}))
}

type rangeExecutorAllocation struct {
	value uint32
	fixed bool
	failure error
}