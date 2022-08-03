package platform_test

import (
	"runtime"
	"testing"

	"github.com/creativeprojects/resticprofile/platform"
	"github.com/stretchr/testify/assert"
)

func TestDarwin(t *testing.T) {
	assert.Equal(t, runtime.GOOS == "darwin", platform.IsDarwin())
}

func TestWindows(t *testing.T) {
	assert.Equal(t, runtime.GOOS == "windows", platform.IsWindows())
}

func TestSupportsSyslog(t *testing.T) {
	assert.Equal(t, runtime.GOOS != "windows", platform.SupportsSyslog())
}
