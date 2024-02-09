//go:build linux

package priority

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartProcessWithIOPriority(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	// Run these 3 tests inside one test, so we don't have concurrency issue
	t.Run("WithNormalIOPriority", func(t *testing.T) {

		output, err := runChildProcess()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(output)
		assert.Contains(t, output, "IOPriority: class = 0, value = 4")

	})

	t.Run("WithBestEffortIOPriority", func(t *testing.T) {

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

	})

	t.Run("WithIdlePriority", func(t *testing.T) {

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

	})
}
