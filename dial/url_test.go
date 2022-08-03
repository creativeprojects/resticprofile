package dial_test

import (
	"testing"

	"github.com/creativeprojects/resticprofile/dial"
	"github.com/stretchr/testify/assert"
)

func TestGetDialAddr(t *testing.T) {
	fixtures := []struct {
		addr     string
		scheme   string
		hostPort string
		isURL    bool
	}{
		// invalid
		{"://", "", "", false},
		// url
		{"scheme://:123", "scheme", ":123", true},
		{"scheme://host:123", "scheme", "host:123", true},
		{"scheme://host", "scheme", "host", true},
		// too short
		{"scheme://", "", "", false},
		{"scheme://:", "", "", false},
		{"c://", "", "", false},
		{"c://:", "", "", false},
		{"c://:123", "", "", false},
		{"c://host:123", "", "", false},
		{"c://host", "", "", false},
		// file
		{"", "", "", false},
		{"//", "", "", false},
		{"file", "", "", false},
		{"/file", "", "", false},
		{"path/file", "", "", false},
		{"/root/file", "", "", false},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.addr, func(t *testing.T) {
			scheme, port, isURL := dial.GetAddr(fixture.addr)

			assert.Equal(t, fixture.isURL, isURL)
			assert.Equal(t, fixture.scheme, scheme)
			assert.Equal(t, fixture.hostPort, port)
		})
	}
}
