// +build ide test_systemd

package fixture_test

import (
	"github.com/akaspin/soil/fixture"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateUnit(t *testing.T) {
	source := `
	# {{.test}}
	[Unit]
	Description=%p
	
	[Service]
	ExecStart=/usr/bin/sleep inf
	
	[Install]
	WantedBy=multi-user.target
	`
	assert.NoError(t, fixture.CreateUnit(
		"/run/systemd/system/test2-one.service",
		source,
		map[string]interface{}{
			"test": 1,
		}))

	assert.NoError(t, fixture.CheckUnitBody(
		"/run/systemd/system/test2-one.service",
		source,
		map[string]interface{}{
			"test": 1,
		}))

	assert.NoError(t, fixture.WaitNoError10(fixture.UnitStatesFn(
		[]string{"test*"},
		map[string]string{
			"test2-one.service": "active",
		},
	)))

	assert.NoError(t, fixture.DestroyUnits("test*"))
}
