package term

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTerminalSingleton(t *testing.T) {
	total := 10
	wg := new(sync.WaitGroup)
	terminals := make([]*Terminal, total)
	for i := range total {
		wg.Go(func() {
			terminals[i] = Get()
		})
	}
	wg.Wait()

	for i := range total {
		assert.NotNil(t, terminals[i])
		assert.Same(t, terminals[0], terminals[i])
	}
}

func TestSetAndGetTerminalSingleton(t *testing.T) {
	terminal := Set(NewTerminal())
	total := 10
	wg := new(sync.WaitGroup)
	terminals := make([]*Terminal, total)
	for i := range total {
		wg.Go(func() {
			terminals[i] = Get()
		})
	}
	wg.Wait()

	for i := range total {
		assert.NotNil(t, terminals[i])
		assert.Same(t, terminal, terminals[i])
	}
}
