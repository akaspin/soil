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
	return r
}

func (r *Range) createFn(id string, config map[string]interface{}, values map[string]string) (res interface{}, err error) {

	// try to find values in already allocated resources
	if allocated, ok := r.allocations[id]; ok && allocated.failure == nil {
		r.log.Tracef(`"id" is already allocated: %d`, allocated.value)
		return nil, nil
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
					return recoveredValue, nil
				}
			} else {
				r.log.Warningf(`recovered value exceeds limits: %s: %d(min) < %d < %d(max)`, id, r.min, recoveredValue, r.max)
			}
		}
	}
	return r.try(id)
}

func (r *Range) updateFn(id string, config map[string]interface{}) (res interface{}, err error) {
	var state rangeExecutorAllocation
	var ok bool
	if state, ok = r.allocations[id]; !ok {
		return nil, fmt.Errorf(`not found: %s`, id)
	}
	if ok && state.failure == nil {
		return nil, fmt.Errorf(`already allocated: %s`, id)
	}
	return r.try(id)
}

func (r *Range) destroyFn(id string) (err error) {
	var state rangeExecutorAllocation
	var ok bool
	if state, ok = r.allocations[id]; !ok {
		return fmt.Errorf(`not found: %s`, id)
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
	return nil
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
		return res, err
	}
	r.notify(id, rangeExecutorAllocation{
		value: res,
	})
	return res, err
}

func (r *Range) allocateBitmap() (res uint32, err error) {
	if ok := r.bitmap.CheckedAdd(r.min); ok {
		return r.min, nil
	}
	iter := r.bitmap.Iterator()
	for iter.HasNext() {
		candidate := iter.Next() + 1
		if candidate > r.max {
			return 0, ErrNotAvailable
		}
		if ok := r.bitmap.CheckedAdd(candidate); ok {
			return candidate, nil
		}
	}
	return 0, ErrNotAvailable
}

func (r *Range) shutdownFn() (err error) {
	return nil
}
