package shutdown

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddAndRunInOrder(t *testing.T) {
	n := 2
	AddHook(func() { n += 5 })
	AddHook(func() { n *= 10 })
	AddHook(nil)
	AddHook(func() { n += 2 })
	AddHook(func() { n *= 2 })

	for i := 0; i < 3; i++ {
		RunHooks()
		assert.Equal(t, 144, n)
	}
	assert.Empty(t, hooks)
}

func TestCanPanicAndRunOthers(t *testing.T) {
	n := 0
	AddHook(func() { n += 1 })
	AddHook(func() { panic(nil) })
	AddHook(func() { n += 1 })

	assert.Panics(t, RunHooks)
	assert.Equal(t, 2, n)
	assert.Empty(t, hooks)
}

func TestContainsHook(t *testing.T) {
	f0 := func() {}
	f1 := func() {}
	f2 := func() {}

	t.Run("empty-tag", func(t *testing.T) {
		RunHooks()
		assert.False(t, ContainsHook())
		AddHook(f0)
		assert.True(t, ContainsHook())
	})

	t.Run("tagged-hook", func(t *testing.T) {
		RunHooks()
		AddHook(f1, "f1", "extra")
		AddHook(f2, "f2")
		assert.False(t, ContainsHook("f1"))
		assert.True(t, ContainsHook("f1", "extra"))
		assert.True(t, ContainsHook("f2"))
	})

	t.Run("re-tagged-hook", func(t *testing.T) {
		RunHooks()
		tags := []string{"f1", "extra"}
		AddHook(f1, tags...)
		assert.True(t, ContainsHook(tags...))
		tags[0] = "f2"
		assert.False(t, ContainsHook(tags...))
	})
}
