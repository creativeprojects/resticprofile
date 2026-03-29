//go:build ssh

package ssh

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

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
			name: "successful connection using RSA key",
			config: Config{
				Host:           "localhost:2222",
				Username:       "resticprofile",
				KnownHostsPath: filepath.Join(tmpDir, "known_hosts"),
				PrivateKeyPath: filepath.Join(tmpDir, "id_rsa"),
				Handler: http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
					resp.Write([]byte("Connection successful using RSA key\n"))
				}),
			},
			connectErr: false,
		},
		{
			name: "successful connection using ECDSA key",
			config: Config{
				Host:           "localhost:2222",
				Username:       "resticprofile",
				KnownHostsPath: filepath.Join(tmpDir, "known_hosts"),
				PrivateKeyPath: filepath.Join(tmpDir, "id_ecdsa"),
				Handler: http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
					resp.Write([]byte("Connection successful using ECDSA key\n"))
				}),
			},
			connectErr: false,
		},
		{
			name: "successful connection using ED25519 key",
			config: Config{
				Host:           "localhost:2222",
				Username:       "resticprofile",
				KnownHostsPath: filepath.Join(tmpDir, "known_hosts"),
				PrivateKeyPath: filepath.Join(tmpDir, "id_ed25519"),
				Handler: http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
					resp.Write([]byte("Connection successful using ED25519 key\n"))
				}),
			},
			connectErr: false,
		},
	}

	for _, fixture := range fixtures {
		for _, client := range []Client{NewOpenSSHClient(fixture.config), NewInternalClient(fixture.config)} {
			t.Run(client.Name()+" "+fixture.name, func(t *testing.T) {
				defer client.Close(context.Background())

				ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
				defer cancel()

				err := client.Connect(ctx)
				if fixture.connectErr {
					require.Error(t, err)
					t.Log(err)
					return
				}
				require.NoError(t, err)

				err = client.Run(ctx, "curl", fmt.Sprintf("http://localhost:%d/", client.TunnelPeerPort()))
				require.NoError(t, err)
			})
		}
	}
}
