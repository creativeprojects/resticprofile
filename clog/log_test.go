package clog

import (
	"testing"
)

func TestLogger(t *testing.T) {
	SetTestLog(t)
	defer ClearTestLog()

	Debug("one", "two", "three")
	Info("one", "two", "three")
	Warning("one", "two", "three")
	Error("one", "two", "three")

	Debugf("%d %d %d", 1, 2, 3)
	Infof("%d %d %d", 1, 2, 3)
	Warningf("%d %d %d", 1, 2, 3)
	Errorf("%d %d %d", 1, 2, 3)
}
