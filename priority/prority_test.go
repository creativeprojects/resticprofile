package priority

import (
	"bytes"
	"io"
	"os/exec"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/platform"
	"github.com/stretchr/testify/assert"
)

func TestStartProcessWithPriority(t *testing.T) {
	t.Run("WithNormalPriority", func(t *testing.T) {
		err := SetClass(Normal)
		if err != nil {
			t.Error(err)
		}

		output, err := runChildProcess()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(output)
		if platform.IsWindows() {
			assert.Contains(t, output, "Priority class: NORMAL")
		} else {
			assert.Contains(t, output, "Process Priority: 0")
		}
	})
	time.Sleep(30 * time.Millisecond)

	t.Run("WithLowerPriority", func(t *testing.T) {
		err := SetClass(Low)
		if err != nil {
			t.Error(err)
		}

		output, err := runChildProcess()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(output)
		if platform.IsWindows() {
			assert.Contains(t, output, "Priority class: BELOW_NORMAL")
		} else {
			assert.Contains(t, output, "Process Priority: 10")
		}
	})
	time.Sleep(30 * time.Millisecond)

	t.Run("WithIdlePriority", func(t *testing.T) {
		err := SetClass(Idle)
		if err != nil {
			t.Error(err)
		}

		output, err := runChildProcess()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(output)
		if platform.IsWindows() {
			assert.Contains(t, output, "Priority class: IDLE")
		} else {
			assert.Contains(t, output, "Process Priority: 19")
		}
	})
	time.Sleep(30 * time.Millisecond)
}

func runChildProcess() (string, error) {
	cmd := exec.Command("go", "run", "./check")
	buffer := &bytes.Buffer{}
	cmd.Stdout = buffer
	cmd.Stderr = buffer
	if err := cmd.Run(); err != nil {
		return "", err
	}
	output, err := io.ReadAll(buffer)
	if err != nil {
		return "", err
	}
	return string(output), nil
}
