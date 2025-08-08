//go:build ssh

package ssh

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSSHClient(t *testing.T) {
	tmpDir := os.Getenv("SSH_TESTS_TMPDIR")
	if tmpDir == "" {
		tmpDir = filepath.Join(os.TempDir(), "resticprofile-ssh-tests")
	}

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
				KnownHostsPath: filepath.Join(tmpDir, "known_hosts"),
			},
			connectErr: true,
		},
		{
			name: "wrong username",
			config: Config{
				Host:           "localhost:2222",
				Username:       "otheruser",
				KnownHostsPath: filepath.Join(tmpDir, "known_hosts"),
				PrivateKeyPath: filepath.Join(tmpDir, "id_rsa"),
			},
			connectErr: true,
		},
		{
			name: "successful connection",
			config: Config{
				Host:           "localhost:2222",
				Username:       "resticprofile",
				KnownHostsPath: filepath.Join(tmpDir, "known_hosts"),
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
