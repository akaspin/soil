// +build !trace,!debug,!notice

package logx_test

import "testing"

func TestLogLevels_NoTags(t *testing.T) {
	checkLevels(t, "INFO", "NOTICE", "WARNING", "ERROR", "CRITICAL")
}
