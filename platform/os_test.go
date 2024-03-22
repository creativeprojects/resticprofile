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
	assert.Equal(t, !platform.IsWindows(), platform.SupportsSyslog())
}

func TestExecutable(t *testing.T) {
	expected := "/path/to/app"
	if platform.IsWindows() {
		expected = "/path/to/app.exe"
	}
	assert.Equal(t, expected, platform.Executable("/path/to/app"))
}
