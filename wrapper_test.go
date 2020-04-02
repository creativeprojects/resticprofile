package main

import (
	"testing"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/stretchr/testify/assert"
)

func TestGetEmptyEnvironment(t *testing.T) {
	profile := config.NewProfile("name")
	wrapper := newResticWrapper("restic", profile, nil)
	env := wrapper.getEnvironment()
	assert.Empty(t, env)
}

func TestGetSingleEnvironment(t *testing.T) {
	profile := config.NewProfile("name")
	profile.Environment = map[string]string{
		"User": "me",
	}
	wrapper := newResticWrapper("restic", profile, nil)
	env := wrapper.getEnvironment()
	assert.Equal(t, []string{"USER=me"}, env)
}

func TestGetMultipleEnvironment(t *testing.T) {
	profile := config.NewProfile("name")
	profile.Environment = map[string]string{
		"User":     "me",
		"Password": "secret",
	}
	wrapper := newResticWrapper("restic", profile, nil)
	env := wrapper.getEnvironment()
	assert.Len(t, env, 2)
	assert.Contains(t, env, "USER=me")
	assert.Contains(t, env, "PASSWORD=secret")
}

func TestEmptyConversionToArgs(t *testing.T) {
	flags := map[string][]string{}
	args := convertIntoArgs(flags)
	assert.Equal(t, []string{}, args)
}

func TestConversionToArgs(t *testing.T) {
	flags := map[string][]string{
		"bool1":  []string{},
		"bool2":  []string{""},
		"int1":   []string{},
		"int2":   []string{"0"},
		"int3":   []string{"-100"},
		"string": []string{"test"},
		"list":   []string{"test1", "test2", "test3"},
	}
	args := convertIntoArgs(flags)
	assert.Len(t, args, 13)
	assert.Contains(t, args, "--bool2")
	assert.Contains(t, args, "\"0\"")
	assert.Contains(t, args, "\"-100\"")
	assert.Contains(t, args, "\"test\"")
	assert.Contains(t, args, "\"test1\"")
	assert.Contains(t, args, "\"test2\"")
	assert.Contains(t, args, "\"test3\"")
}
