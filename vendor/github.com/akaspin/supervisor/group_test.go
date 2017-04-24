package supervisor_test

import (
	"testing"
	"context"
	"github.com/akaspin/supervisor"
	"github.com/stretchr/testify/assert"
	"errors"
	"sync/atomic"
)

type crashable struct {
	*supervisor.Control

	openCn *int64
	closeCn *int64
	doneCn *int64
	waitCn *int64
	err    error
}

func newCrashable(openCn, closeCn, doneCn, waitCn *int64) (c *crashable) {
	c = &crashable{
		Control: supervisor.NewControl(context.TODO()),
		openCn: openCn,
		closeCn: closeCn,
		doneCn: doneCn,
		waitCn: waitCn,
	}
	return
}

func (c *crashable) Open() (err error) {
	go func() {
		<-c.Ctx().Done()
		atomic.AddInt64(c.doneCn, 1)
	}()
	c.Control.Open()
	atomic.AddInt64(c.openCn, 1)
	return
}

func (c *crashable) Close() (err error) {
	c.Control.Close()
	atomic.AddInt64(c.closeCn, 1)
	return
}

func (c *crashable) Wait() (err error) {
	c.Control.Wait()
	atomic.AddInt64(c.waitCn, 1)
	err = c.err
	return
}

func (c *crashable) Crash(err error) {
	c.err = err
	c.Close()
}

func TestGroup_Empty(t *testing.T) {
	g := supervisor.NewGroup(context.TODO())
	g.Open()
	//g.Close()
	g.Wait()
}

func TestGroup_Regular(t *testing.T) {
	var openCn, closeCn, doneCn, waitCn int64
	g := supervisor.NewGroup(
		context.TODO(),
		newCrashable(&openCn, &closeCn, &doneCn, &waitCn),
		newCrashable(&openCn, &closeCn, &doneCn, &waitCn),
		newCrashable(&openCn, &closeCn, &doneCn, &waitCn),
	)
	g.Open()
	g.Close()
	err := g.Wait()
	assert.NoError(t, err)
	assert.Equal(t, []int64{3, 3, 3, 3}, []int64{openCn, closeCn, doneCn, waitCn})
}

func TestGroup_Crash(t *testing.T) {
	var openCn, closeCn, doneCn, waitCn int64
	messy := newCrashable(&openCn, &closeCn, &doneCn, &waitCn)
	g := supervisor.NewGroup(
		context.TODO(),
		newCrashable(&openCn, &closeCn, &doneCn, &waitCn),
		newCrashable(&openCn, &closeCn, &doneCn, &waitCn),
		messy,
	)
	g.Open()
	messy.Crash(errors.New("err"))
	err := g.Wait()
	assert.Error(t, err)
	assert.Equal(t, "err", err.Error())
	assert.Equal(t, []int64{3, 4, 3, 3}, []int64{openCn, closeCn, doneCn, waitCn})
}
