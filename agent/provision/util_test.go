// +build ide test_systemd

package provision_test

import (
	"fmt"
	"github.com/akaspin/soil/fixture"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	podTpl = `
		### SOIL {"Revision":"1.0"}
		### POD {"Name":"{{.Name}}","PodMark":{{.PodMark}},"AgentMark":{{.AgentMark}},"Namespace":"private"}
		{{ $name := .Name -}}
		{{ range $i, $n := .Units -}}
		### UNIT {"Path":"/run/systemd/system/{{$n}}","Create":"start","Update":"restart","Destroy":"stop"}
		{{ end }}
		[Unit]
		Description={{.Name}}
		Before={{ range $i, $n := .Units }}{{ $n }} {{ end }}
		[Service]
		ExecStart=/usr/bin/sleep inf
		[Install]
		WantedBy=multi-user.target`
	unitTpl = `
		[Unit]
		Description={{.Name}}
		[Service]
		ExecStart=/usr/bin/sleep inf
		[Install]
		WantedBy=multi-user.target`
)

func deployPod(name string, units int) (err error) {
	path := "/run/systemd/system/"
	var unitNames []string
	for i := 0; i < units; i++ {
		unitName := fmt.Sprintf("unit-%s-%d.service", name, i)
		unitPath := fmt.Sprintf("%s%s", path, unitName)
		unitNames = append(unitNames, unitName)
		if err = fixture.CreateUnit(unitPath, unitTpl, map[string]interface{}{
			"Name": unitName,
		}); err != nil {
			return err
		}
	}
	podName := fmt.Sprintf("pod-private-%s.service", name)
	podPath := fmt.Sprintf("%s%s", path, podName)
	if err = fixture.CreateUnit(podPath, podTpl, map[string]interface{}{
		"Name":      name,
		"PodMark":   123,
		"AgentMark": 456,
		"Units":     unitNames,
	}); err != nil {
		return err
	}
	return nil
}

func TestDeployPod(t *testing.T) {
	t.Run("2", func(t *testing.T) {
		fixture.DestroyUnits("pod-*", "unit-*")
		defer fixture.DestroyUnits("pod-*", "unit-*")

		assert.NoError(t, deployPod("test", 2))
		assert.NoError(t, fixture.CheckUnitBody("/run/systemd/system/pod-private-test.service", podTpl, map[string]interface{}{
			"Name":      "test",
			"PodMark":   123,
			"AgentMark": 456,
			"Units":     []string{"unit-test-0.service", "unit-test-1.service"},
		}))
		for i := 0; i < 2; i++ {
			assert.NoError(t, fixture.CheckUnitBody(
				fmt.Sprintf("/run/systemd/system/unit-test-%d.service", i),
				unitTpl,
				map[string]interface{}{
					"Name": fmt.Sprintf("unit-test-%d.service", i),
				}))
		}
		assert.NoError(t, fixture.UnitStatesFn(
			[]string{"pod-*", "unit-*"}, map[string]string{
				"unit-test-0.service":      "active",
				"unit-test-1.service":      "active",
				"pod-private-test.service": "active",
			})())
	})
	t.Run("empty", func(t *testing.T) {
		fixture.DestroyUnits("pod-*", "unit-*")
		defer fixture.DestroyUnits("pod-*", "unit-*")

		assert.NoError(t, deployPod("test", 0))
		assert.NoError(t, fixture.CheckUnitBody("/run/systemd/system/pod-private-test.service", podTpl, map[string]interface{}{
			"Name":      "test",
			"PodMark":   123,
			"AgentMark": 456,
			"Units":     []string{},
		}))
		assert.NoError(t, fixture.UnitStatesFn(
			[]string{"pod-*", "unit-*"}, map[string]string{
				"pod-private-test.service": "active",
			})())
		fixture.DestroyUnits("pod-*", "unit-*")
		assert.NoError(t, fixture.UnitStatesFn(
			[]string{"pod-*", "unit-*"}, map[string]string{})())
	})
}
