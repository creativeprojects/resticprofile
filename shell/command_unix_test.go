//go:build !windows

package shell

import (
	"bytes"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInterruptShellCommand(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}

	sigChan := make(chan os.Signal, 1)

	cmd := NewSignalledCommand(mockBinary, []string{"test", "--sleep", "3000"}, sigChan)
	cmd.Stdout = buffer

	// Will ask us to stop in 100ms
	go func() {
		time.Sleep(100 * time.Millisecond)
		sigChan <- syscall.SIGINT
	}()
	start := time.Now()
	_, _, err := cmd.Run()
	require.Error(t, err)

	// check it ran for more than 100ms (but less than 500ms - the build agent can be very slow at times)
	duration := time.Since(start)
	assert.GreaterOrEqual(t, duration.Milliseconds(), int64(100))
	assert.Less(t, duration.Milliseconds(), int64(500))
}
