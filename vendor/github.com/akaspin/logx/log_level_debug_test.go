// +build debug

package logx_test

import "testing"

func TestLogLevels_Debug(t *testing.T) {
	checkLevels(t, "DEBUG", "INFO", "NOTICE", "WARNING", "ERROR", "CRITICAL")
}
