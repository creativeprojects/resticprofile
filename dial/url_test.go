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
		// supported schemes
		{"TCP://:123", "tcp", ":123", true},
		{"UDP://:123", "udp", ":123", true},
		{"tcp://:123", "tcp", ":123", true},
		{"udp://:123", "udp", ":123", true},
		{"syslog://:123", "syslog", ":123", true},
		{"syslog-tcp://:123", "syslog-tcp", ":123", true},
		// url
		{"syslog://:123", "syslog", ":123", true},
		{"syslog://host:123", "syslog", "host:123", true},
		{"syslog://host", "syslog", "host", true},
		{"syslog://", "syslog", "", true},
		{"syslog:", "syslog", "", true},
		// too short
		{"syslog:opaque", "", "", false},
		{"tcp://", "", "", false},
		{"tcp:", "", "", false},
		{"syslog-tcp:", "", "", false},
		{"udp:", "", "", false},
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
		{"temp:/t/backup.log", "", "", false},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.addr, func(t *testing.T) {
			scheme, port, isURL := dial.GetAddr(fixture.addr)

			assert.Equal(t, fixture.isURL, isURL)
			assert.Equal(t, fixture.scheme, scheme)
			assert.Equal(t, fixture.hostPort, port)

			assert.Equal(t, fixture.isURL, dial.IsSupportedURL(fixture.addr))
		})
	}
}

func TestIsUrl(t *testing.T) {
	assert.True(t, dial.IsURL("ftp://"))
	assert.True(t, dial.IsURL("http://"))
	assert.False(t, dial.IsURL("c://"))
	assert.False(t, dial.IsURL("temp:/t/backup.log"))
	assert.False(t, dial.IsURL(""))
}
