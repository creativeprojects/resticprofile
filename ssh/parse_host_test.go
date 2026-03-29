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
			host:     "example.com:220",
			wantHost: "example.com",
			wantPort: 220,
		},
		{
			name:     "IPv4 with port",
			host:     "192.168.1.1:2222",
			wantHost: "192.168.1.1",
			wantPort: 2222,
		},
		{
			name:     "IPv6 with port",
			host:     "[2001:db8::1]:220",
			wantHost: "[2001:db8::1]",
			wantPort: 220,
		},
		{
			name:     "IPv6 without brackets with port",
			host:     "2001:db8::1:220",
			wantHost: "2001:db8::1",
			wantPort: 220,
		},
		{
			name:     "IPv6 without port",
			host:     "2001:db8::1",
			wantHost: "2001:db8::1",
			wantPort: 0,
		},
		{
			name:     "local IPv6 without port",
			host:     "::1",
			wantHost: "::1",
			wantPort: 0,
		},
		{
			name:     "local IPv6 with port",
			host:     "::1:220",
			wantHost: "::1",
			wantPort: 220,
		},
		{
			name:     "host with multiple colons",
			host:     "user:pass@example.com:220",
			wantHost: "user:pass@example.com",
			wantPort: 220,
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
