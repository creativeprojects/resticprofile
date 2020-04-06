//+build linux

package priority

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartProcessWithNormalIOPriority(t *testing.T) {

	output, err := runChildProcess()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(output)
	assert.Contains(t, output, "IOPriority: class = 0, value = 4")

}

func TestStartProcessWithBestEffortIOPriority(t *testing.T) {

	err := SetIONice(2, 6)
	if err != nil {
		t.Fatal(err)
	}

	output, err := runChildProcess()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(output)
	assert.Contains(t, output, "IOPriority: class = 2, value = 6")

}

func TestStartProcessWithIdleIOPriority(t *testing.T) {

	err := SetIONice(3, 4)
	if err != nil {
		t.Fatal(err)
	}

	output, err := runChildProcess()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(output)
	assert.Contains(t, output, "IOPriority: class = 3, value = 4")

}
