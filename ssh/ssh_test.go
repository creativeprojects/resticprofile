//go:build ssh

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

	tmpDir := os.Getenv("KEYS_TMP_DIR")

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
		{
			name: "wrong username",
			config: Config{
				Host:           "localhost:2222",
				Username:       "otheruser",
				KnownHostsPath: filepath.Join(wd, "tests/known_hosts"),
				PrivateKeyPath: filepath.Join(tmpDir, "id_rsa"),
			},
			connectErr: true,
		},
		{
			name: "successful connection",
			config: Config{
				Host:           "localhost:2222",
				Username:       "resticprofile",
				KnownHostsPath: filepath.Join(wd, "tests/known_hosts"),
				PrivateKeyPath: filepath.Join(tmpDir, "id_rsa"),
			},
			connectErr: false,
		},
	}

	for _, fixture := range fixtures {
		for _, client := range []Client{NewOpenSSHClient(fixture.config), NewInternalClient(fixture.config)} {
			t.Run(client.Name()+" "+fixture.name, func(t *testing.T) {
				defer client.Close()

				err := client.Connect()
				if fixture.connectErr {
					require.Error(t, err)
					t.Log(err)
					return
				}
				require.NoError(t, err)
			})
		}
	}
}
