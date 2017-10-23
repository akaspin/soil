// +build ide test_unit

package scheduler_test

import (
	"context"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

type dummyArbiterEntity struct {
	mu       sync.Mutex
	errors   []error
	messages []bus.Message
}

func (e *dummyArbiterEntity) notify(err error, message bus.Message) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.errors = append(e.errors, err)
	e.messages = append(e.messages, message)
}

func (e *dummyArbiterEntity) assertErrors(t *testing.T, errors []error) {
	t.Helper()
	for i, err := range e.errors {
		if err == nil && errors[i] == nil {
			continue
		}
		if fmt.Sprint(err) != fmt.Sprint(errors[i]) {
			t.Errorf("not equal [%d] (expected)%v != (actual)%v", i, errors[i], err)
			t.Fail()
			return
		}
	}
}

func TestArbiter_ConsumeMessage(t *testing.T) {
	arbiter := scheduler.NewArbiter(context.Background(), logx.GetLog("test"), "test",
		scheduler.ArbiterConfig{
			Required: manifest.Constraint{"${drain}": "!= true"},
		},
	)
	entity1 := &dummyArbiterEntity{}
	entity2 := &dummyArbiterEntity{}
	entity3 := &dummyArbiterEntity{}

	assert.NoError(t, arbiter.Open())

	t.Run("register", func(t *testing.T) {
		arbiter.Bind("1", manifest.Constraint{"${1}": "true"}, entity1.notify)
		arbiter.Bind("2", manifest.Constraint{"${2}": "true"}, entity2.notify)
		// assert no actions
		assert.Len(t, entity1.errors, 0)
		assert.Len(t, entity2.errors, 0)
	})
	t.Run("enable 1 and 2", func(t *testing.T) {
		arbiter.ConsumeMessage(bus.NewMessage("", map[string]string{
			"1": "true",
			"2": "true",
		}))
		time.Sleep(time.Millisecond * 100)
		assert.Equal(t, []error{nil}, entity1.errors)
		assert.Equal(t, []error{nil}, entity2.errors)
	})
	t.Run("disable 2", func(t *testing.T) {
		arbiter.ConsumeMessage(bus.NewMessage("", map[string]string{
			"1": "true",
		}))
		time.Sleep(time.Millisecond * 100)
		entity1.assertErrors(t, []error{
			nil,
			nil,
		})
		entity2.assertErrors(t, []error{
			nil,
			fmt.Errorf(`constraint failed: "${2}":"true" ("${2}":"true")`),
		})
	})
	t.Run("drain", func(t *testing.T) {
		arbiter.ConsumeMessage(bus.NewMessage("", map[string]string{
			"drain": "true",
		}))
		time.Sleep(time.Millisecond * 100)
		entity1.assertErrors(t, []error{
			nil,
			nil,
			fmt.Errorf(`constraint failed: "true":"!= true" ("${drain}":"!= true")`),
		})
		entity2.assertErrors(t, []error{
			nil,
			fmt.Errorf(`constraint failed: "${2}":"true" ("${2}":"true")`),
			fmt.Errorf(`constraint failed: "true":"!= true" ("${drain}":"!= true")`),
		})
	})
	t.Run("enable all", func(t *testing.T) {
		arbiter.ConsumeMessage(bus.NewMessage("", map[string]string{
			"1": "true",
			"2": "true",
		}))
		time.Sleep(time.Millisecond * 100)
		entity1.assertErrors(t, []error{
			nil,
			nil,
			fmt.Errorf(`constraint failed: "true":"!= true" ("${drain}":"!= true")`),
			nil,
		})
		entity2.assertErrors(t, []error{
			nil,
			fmt.Errorf(`constraint failed: "${2}":"true" ("${2}":"true")`),
			fmt.Errorf(`constraint failed: "true":"!= true" ("${drain}":"!= true")`),
			nil,
		})
	})
	t.Run("register 3", func(t *testing.T) {
		arbiter.Bind("3", manifest.Constraint{"${2}": "true"}, entity3.notify)
		time.Sleep(time.Millisecond * 100)
		entity1.assertErrors(t, []error{
			nil,
			nil,
			fmt.Errorf(`constraint failed: "true":"!= true" ("${drain}":"!= true")`),
			nil,
		})
		entity2.assertErrors(t, []error{
			nil,
			fmt.Errorf(`constraint failed: "${2}":"true" ("${2}":"true")`),
			fmt.Errorf(`constraint failed: "true":"!= true" ("${drain}":"!= true")`),
			nil,
		})
		entity3.assertErrors(t, []error{
			nil,
		})
	})

	arbiter.Close()
	arbiter.Wait()
}
