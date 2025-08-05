package main

import (
	"net/http/httptest"
	"testing"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/stretchr/testify/assert"
)

func TestSendRemoteFiles(t *testing.T) {
	recorder := httptest.NewRecorder()
	sendRemoteFiles(&config.Remote{
		ConfigurationFile: "test_config.json",
		ProfileName:       "test_profile",
	}, "test_remote", []string{"arg1", "arg2"}, recorder)
	assert.Equal(t, recorder.Code, 200)
	assert.Equal(t, recorder.Header().Get("Content-Type"), "application/x-tar")
}
