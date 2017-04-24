package logx_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"unsafe"
)

type outputMock struct{}

type atomicMock struct {
	structVPtr *unsafe.Pointer
	int32VPtr  *int32
	stringVPtr *unsafe.Pointer
}

func newAtomicMock() *atomicMock {
	return &atomicMock{
		new(unsafe.Pointer),
		new(int32),
		new(unsafe.Pointer),
	}
}

type lockerMock struct {
	rwMu *sync.RWMutex
	mu   *sync.Mutex

	structV *outputMock
	int32V  int32
	stringV string
}

func newLockerMock() *lockerMock {
	return &lockerMock{
		&sync.RWMutex{},
		&sync.Mutex{},
		&outputMock{},
		0,
		"",
	}
}

func Benchmark_AtomicSingleSerial(b *testing.B) {
	m := newAtomicMock()

	for i := 0; i < b.N; i++ {
		atomic.StorePointer(m.structVPtr, (unsafe.Pointer)(&outputMock{}))
		for j := 0; j < 100; j++ {
			_ = (*outputMock)(atomic.LoadPointer(m.structVPtr))
		}
	}
}

func Benchmark_AtomicMultipleSerial(b *testing.B) {
	m := newAtomicMock()

	for i := 0; i < b.N; i++ {
		atomic.StorePointer(m.structVPtr, (unsafe.Pointer)(&outputMock{}))
		atomic.StoreInt32(m.int32VPtr, int32(0))
		atomic.StorePointer(m.stringVPtr, (unsafe.Pointer)(new(string)))
		for j := 0; j < 100; j++ {
			_ = (*outputMock)(atomic.LoadPointer(m.structVPtr))
			_ = atomic.LoadInt32(m.int32VPtr)
			_ = (*string)(atomic.LoadPointer(m.stringVPtr))
		}
	}
}

func Benchmark_AtomicSingleParallel(b *testing.B) {
	m := newAtomicMock()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			atomic.StorePointer(m.structVPtr, (unsafe.Pointer)(&outputMock{}))
			for j := 0; j < 100; j++ {
				_ = (*outputMock)(atomic.LoadPointer(m.structVPtr))
			}
		}
	})
}

func Benchmark_AtomicMultipleParallel(b *testing.B) {
	m := newAtomicMock()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			atomic.StorePointer(m.structVPtr, (unsafe.Pointer)(&outputMock{}))
			atomic.StoreInt32(m.int32VPtr, int32(0))
			atomic.StorePointer(m.stringVPtr, (unsafe.Pointer)(new(string)))
			for j := 0; j < 100; j++ {
				_ = (*outputMock)(atomic.LoadPointer(m.structVPtr))
				_ = atomic.LoadInt32(m.int32VPtr)
				_ = (*string)(atomic.LoadPointer(m.stringVPtr))
			}
		}
	})
}

func Benchmark_LockRWMutexSerial(b *testing.B) {
	m := newLockerMock()

	for i := 0; i < b.N; i++ {
		m.rwMu.Lock()
		m.structV = &outputMock{}
		m.rwMu.Unlock()

		for j := 0; j < 100; j++ {
			m.rwMu.RLock()
			_ = m.structV
			m.rwMu.RUnlock()
		}
	}
}

func Benchmark_LockMutexSerial(b *testing.B) {
	m := newLockerMock()

	for i := 0; i < b.N; i++ {
		m.mu.Lock()
		m.structV = &outputMock{}
		m.mu.Unlock()

		for j := 0; j < 100; j++ {
			m.mu.Lock()
			_ = m.structV
			m.mu.Unlock()
		}
	}
}

func Benchmark_LockRWMutexParallel(b *testing.B) {
	m := newLockerMock()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.rwMu.Lock()
			m.structV = &outputMock{}
			m.rwMu.Unlock()

			for j := 0; j < 100; j++ {
				m.rwMu.RLock()
				_ = m.structV
				m.rwMu.RUnlock()
			}
		}
	})
}

func Benchmark_LockMutexParallel(b *testing.B) {
	m := newLockerMock()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.mu.Lock()
			m.structV = &outputMock{}
			m.mu.Unlock()

			for j := 0; j < 100; j++ {
				m.mu.Lock()
				_ = m.structV
				m.mu.Unlock()
			}
		}
	})
}
