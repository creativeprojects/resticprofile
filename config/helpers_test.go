package config

import (
	"bytes"
)

// Helpers for tests

func getProfile(configFormat, configString, profileKey string) (*Profile, error) {
	c, err := Load(bytes.NewBufferString(configString), configFormat)
	if err != nil {
		return nil, err
	}

	profile, err := c.getProfile(profileKey)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

func getResolvedProfile(configFormat, configString, profileKey string) (*Profile, error) {
	c, err := Load(bytes.NewBufferString(configString), configFormat)
	if err != nil {
		return nil, err
	}

	profile, err := c.GetProfile(profileKey)
	if err != nil {
		return nil, err
	}

	return profile, nil
}
