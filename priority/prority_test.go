package priority

import (
	"bytes"
	"io/ioutil"
	"os/exec"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartProcessWithNormalPriority(t *testing.T) {
	err := SetClass(Normal)
	if err != nil {
		t.Error(err)
	}

	output, err := runChildProcess()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(output)
	if runtime.GOOS == "windows" {
		assert.Contains(t, output, "Priority class: NORMAL")
	} else {
		assert.Contains(t, output, "Priority: 0")
	}
}

func TestStartProcessWithLowerPriority(t *testing.T) {
	err := SetClass(Low)
	if err != nil {
		t.Error(err)
	}

	output, err := runChildProcess()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(output)
	if runtime.GOOS == "windows" {
		assert.Contains(t, output, "Priority class: BELOW_NORMAL")
	} else {
		assert.Contains(t, output, "Priority: 10")
	}
}

func TestStartProcessWithBackgroundPriority(t *testing.T) {
	err := SetClass(Background)
	if err != nil {
		t.Error(err)
	}

	output, err := runChildProcess()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(output)
	if runtime.GOOS == "windows" {
		assert.Contains(t, output, "Priority class: IDLE")
	} else {
		assert.Contains(t, output, "Priority: 15")
	}
}

func runChildProcess() (string, error) {
	cmd := exec.Command("go", "run", "./check")
	buffer := &bytes.Buffer{}
	cmd.Stdout = buffer
	cmd.Stderr = buffer
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	output, err := ioutil.ReadAll(buffer)
	if err != nil {
		return "", err
	}
	return string(output), nil
}
