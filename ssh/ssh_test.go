package ssh

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSSHClient(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	t.Logf("Current working directory: %s", wd)

	fixtures := []struct {
		name       string
		config     Config
		connectErr bool
	}{
		{
			name: "no public key",
			config: Config{
				Host:           "localhost:2222",
				Username:       "resticprofile",
				KnownHostsPath: filepath.Join(wd, "./tests/known_hosts"),
			},
			connectErr: true,
		},
	}

	for _, fixture := range fixtures {
		for _, client := range []Client{NewOpenSSHClient(fixture.config), NewInternalClient(fixture.config)} {
			t.Run(client.Name()+" "+fixture.name, func(t *testing.T) {
				err := client.Connect()
				if fixture.connectErr {
					require.Error(t, err, "expected error for config: %v", fixture.config)
					t.Log(err)
					return
				}
				require.NoError(t, err, "unexpected error for config: %v", fixture.config)
			})
		}
	}
}
