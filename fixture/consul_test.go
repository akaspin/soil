// +build ide test_cluster

package fixture_test

import (
	"github.com/akaspin/soil/fixture"
	"testing"
	"time"
)

func TestConsulServer(t *testing.T) {
	t.Run("0 with space", func(t *testing.T) {
		s := fixture.NewConsulServer(t, nil)
		defer s.Clean()
		s.Up()
		time.Sleep(time.Second)
		s.Down()
	})
}
