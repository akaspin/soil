// +build trace

package logx_test

import "testing"

func TestLogLevels_Trace(t *testing.T) {
	checkLevels(t, "TRACE", "DEBUG", "INFO", "NOTICE", "WARNING", "ERROR", "CRITICAL")
}
