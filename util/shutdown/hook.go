package shutdown

import (
	"sync"

	"golang.org/x/exp/slices"
)

var (
	lock  sync.Mutex
	hooks []hookFunc
)

type hookFunc struct {
	fn   func()
	tags []string
}

// AddHook add a func to run when RunHooks is invoked. Hooks are run in FIFO order
func AddHook(hook func(), tag ...string) {
	if hook != nil {
		lock.Lock()
		defer lock.Unlock()

		hooks = append(hooks, hookFunc{fn: hook, tags: append([]string{}, tag...)})
	}
}

// ContainsHook returns true if a hook with the specified hook tag was added
func ContainsHook(tag ...string) bool {
	lock.Lock()
	defer lock.Unlock()

	return slices.ContainsFunc(hooks, func(fn hookFunc) bool {
		return slices.Equal(fn.tags, tag)
	})
}

// RunHooks runs all hooks and clears the hooks list
func RunHooks() {
	lock.Lock()
	defer lock.Unlock()

	runHook := func(hook func()) { hook() }

	// Running hooks with defer to ensure all are started even if one panics
	// Hooks are started in FIFO order but defer is LIFO, reverse for loop is needed
	for i := len(hooks); i > 0; i-- {
		defer runHook(hooks[i-1].fn)
	}
	hooks = nil
}
