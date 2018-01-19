package supervisor_test

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

const compositeTestIterations = 50

type testingWatcher struct {
	in  chan string
	wg  sync.WaitGroup
	res []string
}

func newTestingWatcher(expect int, components ...*testingComponent) (w *testingWatcher) {
	w = &testingWatcher{
		in: make(chan string),
	}
	for _, comp := range components {
		comp.reportChan = w.in
	}
	w.wg.Add(expect)
	go func() {
		for m := range w.in {
			//println(m)
			w.res = append(w.res, m)
			w.wg.Done()
		}
	}()
	return
}

type testingComponent struct {
	name     string
	errOpen  error
	errClose error
	errWait  error

	eventsMu   sync.Mutex
	events     []string
	closedChan chan struct{}
	reportChan chan string
}

func newTestingComponent(name string, errOpen, errClose, errWait error) (c *testingComponent) {
	c = &testingComponent{
		name:       name,
		errOpen:    errOpen,
		errClose:   errClose,
		errWait:    errWait,
		closedChan: make(chan struct{}),
	}
	return
}

func (c *testingComponent) appendEvent(e string) {
	c.eventsMu.Lock()
	defer c.eventsMu.Unlock()
	c.events = append(c.events, e)
	if c.reportChan != nil {
		c.reportChan <- c.name + "-" + e
	}
}

func (c *testingComponent) assertEvents(t *testing.T, events ...string) {
	t.Helper()
	c.eventsMu.Lock()
	defer c.eventsMu.Unlock()
	assert.Equal(t, events, c.events, c.name)
}

func (c *testingComponent) assertCycle(t *testing.T) {
	t.Helper()
	c.assertEvents(t, "open", "close", "done")
}

func (c *testingComponent) Open() (err error) {
	err = c.errOpen
	c.appendEvent("open")
	//println("o", c.name)
	return
}

func (c *testingComponent) Close() (err error) {
	err = c.errClose
	c.appendEvent("close")
	close(c.closedChan)
	//println("c", c.name)
	return
}

func (c *testingComponent) Wait() (err error) {
	err = c.errWait
	<-c.closedChan
	c.appendEvent("done")
	//println("w", c.name)
	return
}
