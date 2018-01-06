package estimator

import (
	"fmt"
	"github.com/RoaringBitmap/roaring"
	"github.com/akaspin/soil/manifest"
	"strconv"
)

type rangeExecutorAllocation struct {
	value   uint32
	failure error
}

type Range struct {
	*base
	min uint32
	max uint32

	bitmap      *roaring.Bitmap
	allocations map[string]rangeExecutorAllocation // allocation requests by id
}

func NewRange(globalConfig GlobalConfig, config Config) (r *Range) {
	r = &Range{
		bitmap:      roaring.New(),
		allocations: map[string]rangeExecutorAllocation{},
	}
	if v, ok := config.Provider.Config["min"]; ok {
		r.min = uint32(v.(int))
	}
	if v, ok := config.Provider.Config["max"]; ok {
		r.max = uint32(v.(int))
	}
	r.base = newBase(globalConfig, config, r)
	return
}

func (r *Range) createFn(id string, config map[string]interface{}, values map[string]string) (res interface{}, err error) {

	// try to find values in already allocated resources
	if allocated, ok := r.allocations[id]; ok && allocated.failure == nil {
		r.log.Tracef(`"id" is already allocated: %d`, allocated.value)
		return
	}

	// try to find recovered value
	var recoveredValue uint32

	if raw, ok := values["value"]; ok {
		if parsed, parseErr := strconv.ParseUint(raw, 10, 32); parseErr != nil {
			r.log.Warningf(`can't parse value: %s:%s`, id, raw)
		} else {
			if recoveredValue = uint32(parsed); recoveredValue >= r.min && recoveredValue <= r.max {
				if ok = r.bitmap.CheckedAdd(recoveredValue); ok {
					r.log.Tracef(`"%s" allocated from recovery: %d`, id, recoveredValue)
					r.notify(id, rangeExecutorAllocation{
						value: recoveredValue,
					})
					res = recoveredValue
					return
				}
			} else {
				r.log.Warningf(`recovered value exceeds limits: %s: %d(min) < %d < %d(max)`, id, r.min, recoveredValue, r.max)
			}
		}
	}
	res, err = r.try(id)
	return
}

func (r *Range) updateFn(id string, config map[string]interface{}) (res interface{}, err error) {
	var state rangeExecutorAllocation
	var ok bool
	if state, ok = r.allocations[id]; !ok {
		err = fmt.Errorf(`not found: %s`, id)
		return
	}
	if ok && state.failure == nil {
		err = fmt.Errorf(`already allocated: %s`, id)
		return
	}
	res, err = r.try(id)
	return
}

func (r *Range) destroyFn(id string) (err error) {
	var state rangeExecutorAllocation
	var ok bool
	if state, ok = r.allocations[id]; !ok {
		err = fmt.Errorf(`not found: %s`, id)
		return
	}

	if state.failure == nil {
		r.bitmap.Remove(state.value)
	}
	delete(r.allocations, id)
	r.log.Tracef(`deallocated: %s: %v`, id, state)
	r.send(id, nil, nil)

	for allocatedId, alloc := range r.allocations {
		if alloc.failure != nil {
			var res uint32
			var reallocErr error
			if res, reallocErr = r.try(allocatedId); reallocErr != nil {
				r.log.Warningf(`fail to reallocate "%s": %v`, id, reallocErr)
				continue
			}
			r.log.Infof(`reallocated %s: %v`, allocatedId, res)
		}
	}
	return
}

func (r *Range) notify(id string, alloc rangeExecutorAllocation) {
	r.allocations[id] = alloc
	r.send(id, alloc.failure, manifest.FlatMap{
		"value": fmt.Sprintf("%d", alloc.value),
	})
	r.log.Debugf(`downstream notified: %s:%v`, id, alloc)
}

func (r *Range) try(id string) (res uint32, err error) {
	res, err = r.allocateBitmap()
	if err != nil {
		r.notify(id, rangeExecutorAllocation{
			failure: err,
		})
		return
	}
	r.notify(id, rangeExecutorAllocation{
		value: res,
	})
	return
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
			err = ErrNotAvailable
			return
		}
		if ok := r.bitmap.CheckedAdd(candidate); ok {
			res = candidate
			return
		}
	}
	err = ErrNotAvailable
	return
}

func (r *Range) shutdownFn() (err error) {
	return
}
