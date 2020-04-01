package config

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/spf13/viper"
)

func TestNoProfile(t *testing.T) {
	testConfig := ""
	profile, err := getProfile(testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.Nil(t, profile)
}

func TestEmptyProfile(t *testing.T) {
	testConfig := `[profile]
`
	profile, err := getProfile(testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, profile)
	assert.Equal(t, "profile", profile.Name)
}

func TestNoInitializeValue(t *testing.T) {
	testConfig := `[profile]
`
	profile, err := getProfile(testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, profile)
	assert.Equal(t, false, profile.Initialize)
}

func TestInitializeValueFalse(t *testing.T) {
	testConfig := `[profile]
initialize = false
`
	profile, err := getProfile(testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, profile)
	assert.Equal(t, false, profile.Initialize)
}

func TestInitializeValueTrue(t *testing.T) {
	testConfig := `[profile]
initialize = true
`
	profile, err := getProfile(testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, profile)
	assert.Equal(t, true, profile.Initialize)
}

func TestInheritedInitializeValueTrue(t *testing.T) {
	testConfig := `[parent]
initialize = true

[profile]
inherit = "parent"
`
	profile, err := getProfile(testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, profile)
	assert.Equal(t, true, profile.Initialize)
}

func TestOverriddenInitializeValueFalse(t *testing.T) {
	testConfig := `[parent]
initialize = true

[profile]
initialize = false
inherit = "parent"
`
	profile, err := getProfile(testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, profile)
	assert.Equal(t, false, profile.Initialize)
}

func TestMultiInheritance(t *testing.T) {
	testConfig := `
[grand-parent]
repository = "grand-parent"
first-value = 1
override-value = 1

[parent]
inherit = "grand-parent"
initialize = true
repository = "parent"
second-value = 2
override-value = 2
quiet = true

[profile]
inherit = "parent"
third-value = 3
verbose = true
quiet = false
`
	profile, err := getProfile(testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, profile)
	assert.Equal(t, "profile", profile.Name)
	assert.Equal(t, "parent", profile.Repository)
	assert.Equal(t, true, profile.Initialize)
	assert.Equal(t, int64(1), profile.OtherFlags["first-value"])
	assert.Equal(t, int64(2), profile.OtherFlags["second-value"])
	assert.Equal(t, int64(3), profile.OtherFlags["third-value"])
	assert.Equal(t, int64(2), profile.OtherFlags["override-value"])
	assert.Equal(t, false, profile.Quiet)
	assert.Equal(t, true, profile.Verbose)
}

func getProfile(configString, profileKey string) (*Profile, error) {
	viper.SetConfigType("toml")
	err := viper.ReadConfig(bytes.NewBufferString(configString))
	if err != nil {
		return nil, err
	}

	profile, err := LoadProfile(profileKey)
	if err != nil {
		return nil, err
	}
	return profile, nil
}
