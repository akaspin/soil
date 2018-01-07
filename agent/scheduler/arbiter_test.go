// +build ide test_unit

package scheduler_test

import (
	"context"
	"fmt"
	"github.com/akaspin/logx"
	"github.com/akaspin/soil/agent/bus"
	"github.com/akaspin/soil/agent/scheduler"
	"github.com/akaspin/soil/fixture"
	"github.com/akaspin/soil/manifest"
	"github.com/stretchr/testify/assert"
	"reflect"
	"regexp"
	"strconv"
	"sync"
	"testing"
)

func TestRegex(t *testing.T) {
	expr := regexp.MustCompile(`^resource\..+\.allocated$`)
	var res []string
	src := []string{
		"pod.1.allocated",
		"resource.allocated",
		"resource.test.1.allocated",
		"resource.test.1.value",
	}
	for _, s := range src {
		if !expr.Match([]byte(s)) {
			res = append(res, s)
		}
	}
	assert.Equal(t, []string{
		"pod.1.allocated",
		"resource.allocated",
		"resource.test.1.value",
	}, res)
}

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

func (e *dummyArbiterEntity) checkErrorsFn(errs ...error) func() (err error) {
	return func() (err error) {
		e.mu.Lock()
		defer e.mu.Unlock()
		for i, err1 := range e.errors {
			if err1 == nil && errs[i] == nil {
				continue
			}
			if fmt.Sprint(err1) != fmt.Sprint(errs[i]) {
				err = fmt.Errorf("not equal [%d] (expected)%v != (actual)%v", i, errs[i], err)
				return
			}
		}
		return
	}
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
	for i := 0; i < 5; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			arbiter := scheduler.NewArbiter(context.Background(), logx.GetLog("test"), "test",
				scheduler.ArbiterConfig{
					Required: manifest.Constraint{"${drain}": "!= true"},
					ConstraintOnly: []*regexp.Regexp{
						regexp.MustCompile(`^status\.pod\..+`),
					},
				},
			)
			entity1 := &dummyArbiterEntity{}
			entity2 := &dummyArbiterEntity{}
			entity3 := &dummyArbiterEntity{}

			assert.NoError(t, arbiter.Open())

			t.Run("bind 1 and 2", func(t *testing.T) {
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
				fixture.WaitNoError10(t, entity1.checkErrorsFn(nil))
				fixture.WaitNoError10(t, entity2.checkErrorsFn(nil))
			})
			t.Run("disable 2", func(t *testing.T) {
				arbiter.ConsumeMessage(bus.NewMessage("", map[string]string{
					"1": "true",
				}))
				fixture.WaitNoError10(t, entity1.checkErrorsFn(nil, nil))
				fixture.WaitNoError10(t, entity2.checkErrorsFn(
					nil,
					fmt.Errorf(`constraint failed: "${2}":"true" ("${2}":"true")`),
				))
			})
			t.Run("drain on", func(t *testing.T) {
				arbiter.ConsumeMessage(bus.NewMessage("", map[string]string{
					"drain": "true",
				}))
				fixture.WaitNoError10(t, entity1.checkErrorsFn(
					nil,
					nil,
					fmt.Errorf(`constraint failed: "true":"!= true" ("${drain}":"!= true")`),
				))
				fixture.WaitNoError10(t, entity2.checkErrorsFn(
					nil,
					fmt.Errorf(`constraint failed: "${2}":"true" ("${2}":"true")`),
					fmt.Errorf(`constraint failed: "true":"!= true" ("${drain}":"!= true")`),
				))
			})
			t.Run("drain off", func(t *testing.T) {
				arbiter.ConsumeMessage(bus.NewMessage("", map[string]string{
					"1": "true",
					"2": "true",
				}))
				fixture.WaitNoError10(t, entity1.checkErrorsFn(
					nil,
					nil,
					fmt.Errorf(`constraint failed: "true":"!= true" ("${drain}":"!= true")`),
					nil,
				))
				fixture.WaitNoError10(t, entity2.checkErrorsFn(
					nil,
					fmt.Errorf(`constraint failed: "${2}":"true" ("${2}":"true")`),
					fmt.Errorf(`constraint failed: "true":"!= true" ("${drain}":"!= true")`),
					nil,
				))
			})
			t.Run("bind 3", func(t *testing.T) {
				arbiter.Bind("3", manifest.Constraint{"${2}": "true"}, entity3.notify)
				fixture.WaitNoError10(t, entity1.checkErrorsFn(
					nil,
					nil,
					fmt.Errorf(`constraint failed: "true":"!= true" ("${drain}":"!= true")`),
					nil,
				))
				fixture.WaitNoError10(t, entity2.checkErrorsFn(
					nil,
					fmt.Errorf(`constraint failed: "${2}":"true" ("${2}":"true")`),
					fmt.Errorf(`constraint failed: "true":"!= true" ("${drain}":"!= true")`),
					nil,
				))
				fixture.WaitNoError10(t, entity3.checkErrorsFn(
					nil,
				))
			})
			t.Run("unbind 1", func(t *testing.T) {
				arbiter.Unbind("1", func() {
					entity1.notify(fmt.Errorf("unbind"), bus.NewMessage("", nil))
				})
				fixture.WaitNoError10(t, entity1.checkErrorsFn(
					nil,
					nil,
					fmt.Errorf(`constraint failed: "true":"!= true" ("${drain}":"!= true")`),
					nil,
					fmt.Errorf("unbind"),
				))
			})
			t.Run("bind 1 with status constraint", func(t *testing.T) {
				arbiter.Bind("1", manifest.Constraint{
					"${1}":            "true",
					"${status.pod.1}": "ok",
				}, entity1.notify)
				fixture.WaitNoError10(t, entity1.checkErrorsFn(
					nil,
					nil,
					fmt.Errorf(`constraint failed: "true":"!= true" ("${drain}":"!= true")`),
					nil,
					fmt.Errorf("unbind"),
					fmt.Errorf("constraint failed: \"${status.pod.1}\":\"ok\" (\"${status.pod.1}\":\"ok\")"),
				))
			})
			t.Run("update with constraintOnly", func(t *testing.T) {
				arbiter.ConsumeMessage(bus.NewMessage("private", map[string]string{
					"1":            "true",
					"2":            "true",
					"3":            "true",
					"status.pod.1": "ok",
				}))
				fixture.WaitNoError10(t, entity1.checkErrorsFn(
					nil,
					nil,
					fmt.Errorf(`constraint failed: "true":"!= true" ("${drain}":"!= true")`),
					nil,
					fmt.Errorf("unbind"),
					fmt.Errorf("constraint failed: \"${status.pod.1}\":\"ok\" (\"${status.pod.1}\":\"ok\")"),
					nil,
				))
				fixture.WaitNoError10(t, func() (err error) {
					expect := map[string]string{
						"1": "true",
						"2": "true",
						"3": "true",
					}
					entity1.mu.Lock()
					defer entity1.mu.Unlock()
					lastMsg := entity1.messages[len(entity1.messages)-1]
					var chunk map[string]string
					if err = lastMsg.Payload().Unmarshal(&chunk); err != nil {
						return
					}
					if !reflect.DeepEqual(expect, chunk) {
						err = fmt.Errorf(`%v != %v`, expect, chunk)
					}
					return
				})
			})

			arbiter.Close()
			arbiter.Wait()
		})
	}

}
