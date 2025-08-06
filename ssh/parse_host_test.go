package ssh

import (
	"testing"
)

func TestParseHost(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		wantHost string
		wantPort int
	}{
		{
			name:     "host only",
			host:     "example.com",
			wantHost: "example.com",
			wantPort: 0,
		},
		{
			name:     "host with port",
			host:     "example.com:22",
			wantHost: "example.com",
			wantPort: 22,
		},
		{
			name:     "IPv4 with port",
			host:     "192.168.1.1:2222",
			wantHost: "192.168.1.1",
			wantPort: 2222,
		},
		{
			name:     "IPv6 with port",
			host:     "[2001:db8::1]:22",
			wantHost: "[2001:db8::1]",
			wantPort: 22,
		},
		{
			name:     "IPv6 without brackets with port",
			host:     "2001:db8::1:22",
			wantHost: "2001:db8::1",
			wantPort: 22,
		},
		{
			name:     "host with multiple colons",
			host:     "user:pass@example.com:22",
			wantHost: "user:pass@example.com",
			wantPort: 22,
		},
		{
			name:     "invalid port",
			host:     "example.com:abc",
			wantHost: "example.com",
			wantPort: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHost, gotPort := parseHost(tt.host)
			if gotHost != tt.wantHost {
				t.Errorf("parseHost() host = %v, want %v", gotHost, tt.wantHost)
			}
			if gotPort != tt.wantPort {
				t.Errorf("parseHost() port = %v, want %v", gotPort, tt.wantPort)
			}
		})
	}
}
