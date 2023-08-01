package util

import (
	"os"
	"strings"
	"testing"

	"github.com/creativeprojects/resticprofile/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCanReadOsEnv(t *testing.T) {
	env := NewDefaultEnvironment(os.Environ()...)

	// All values are included
	for _, value := range os.Environ() {
		kv := strings.SplitN(value, "=", 2)
		assert.Equal(t, kv[1], env.Get(strings.TrimSpace(kv[0])))
	}

	// Elements are retained like specified
	assert.ElementsMatch(t, env.Values(), os.Environ())
}

func TestEnvironmentPreservesCase(t *testing.T) {
	assert.Equal(t, !platform.IsWindows(), EnvironmentPreservesCase())
	assert.Equal(t, !platform.IsWindows(), NewDefaultEnvironment().preserveCase)
	assert.False(t, NewFoldingEnvironment().preserveCase)
}

func TestCanSetAndRemove(t *testing.T) {
	env := NewDefaultEnvironment()
	env.Put("Name", "value")
	assert.Equal(t, "value", env.Get("Name"))

	env.Put("N=V", "value")
	assert.Equal(t, "", env.Get("N=V"))
	assert.False(t, env.Has("N=V"))

	env.Put("Name", "")
	assert.Equal(t, "", env.Get("Name"))
	assert.False(t, env.Has("Name"))
}

func TestCaseFolding(t *testing.T) {
	env := NewEnvironment(true)
	foldingEnv := NewEnvironment(false)

	env.Put("Name", "abc")
	foldingEnv.Put("Name", "abc")

	t.Run("Values", func(t *testing.T) {
		values := []string{"Name=abc"}
		assert.Equal(t, values, env.Values())
		assert.Equal(t, values, foldingEnv.Values())
	})

	t.Run("ValuesAsMap", func(t *testing.T) {
		values := map[string]string{"Name": "abc"}
		assert.Equal(t, values, env.ValuesAsMap())
		assert.Equal(t, values, foldingEnv.ValuesAsMap())
	})

	t.Run("Names", func(t *testing.T) {
		names := []string{"Name"}
		foldedNames := []string{"NAME"}
		assert.Equal(t, names, env.Names())
		assert.Equal(t, names, foldingEnv.Names())
		assert.Equal(t, names, env.FoldedNames())
		assert.Equal(t, foldedNames, foldingEnv.FoldedNames())
	})

	t.Run("Get", func(t *testing.T) {
		assert.Equal(t, env.Get("Name"), foldingEnv.Get("Name"))

		assert.Equal(t, "", env.Get("NAME"))
		assert.False(t, env.Has("NAME"))
		assert.Equal(t, "abc", foldingEnv.Get("NAME"))
		assert.True(t, foldingEnv.Has("NAME"))
	})

	t.Run("ResolveName", func(t *testing.T) {
		assert.Equal(t, "NAME", env.ResolveName("NAME"))
		assert.Equal(t, "Name", foldingEnv.ResolveName("NAME"))
	})

	t.Run("Remove", func(t *testing.T) {
		env.Put("ToRemove", "x")
		env.Remove("TOREMOVE")
		foldingEnv.Put("ToRemove", "x")
		foldingEnv.Remove("TOREMOVE")

		assert.True(t, env.Has("ToRemove"))
		assert.False(t, foldingEnv.Has("ToRemove"))

		env.Remove("ToRemove")
		require.False(t, env.Has("ToRemove"))
	})
}
