//+build darwin

package shell

import (
	"bytes"
	"os"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TODO try to make this test pass under Linux
func TestInterruptShellCommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Test not running under Windows")
	}
	buffer := &bytes.Buffer{}

	sigChan := make(chan os.Signal, 1)

	cmd := NewSignalledCommand("sleep", []string{"3"}, sigChan)
	cmd.Stdout = buffer

	// Will ask us to stop in 100ms
	go func() {
		time.Sleep(100 * time.Millisecond)
		sigChan <- syscall.Signal(syscall.SIGINT)
	}()
	start := time.Now()
	err := cmd.Run()
	if err != nil && err.Error() != "signal: interrupt" {
		t.Fatal(err)
	}

	assert.WithinDuration(t, time.Now(), start, 1*time.Second)
}
