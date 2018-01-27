// +build notice

package logx_test

import "testing"

func TestLogLevels_Quiet(t *testing.T) {
	checkLevels(t, "NOTICE", "WARNING", "ERROR", "CRITICAL")
}
