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

	"github.com/creativeprojects/clog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSHClient(t *testing.T) {
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

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
				Host:           "localhost",
				Port:           2222,
				Username:       "resticprofile",
				KnownHostsPath: filepath.Join(tmpDir, "known_hosts"),
			},
			connectErr: true,
		},
		{
			name: "wrong username",
			config: Config{
				Host:            "localhost",
				Port:            2222,
				Username:        "otheruser",
				KnownHostsPath:  filepath.Join(tmpDir, "known_hosts"),
				PrivateKeyPaths: []string{filepath.Join(tmpDir, "id_rsa")},
			},
			connectErr: true,
		},
		// {
		// 	name: "invalid known hosts file",
		// 	config: Config{
		// 		Host:            "localhost",
		// 		Port:            2222,
		// 		Username:        "resticprofile",
		// 		KnownHostsPath:  filepath.Join(tmpDir, "file-not-found"),
		// 		PrivateKeyPaths: []string{filepath.Join(tmpDir, "id_rsa")},
		// 	},
		// 	connectErr: true,
		// },
		{
			name: "successful connection using RSA key",
			config: Config{
				Host:            "localhost",
				Port:            2222,
				Username:        "resticprofile",
				KnownHostsPath:  filepath.Join(tmpDir, "known_hosts"),
				PrivateKeyPaths: []string{filepath.Join(tmpDir, "id_rsa")},
				Handler: http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
					resp.Write([]byte("Connection successful using RSA key\n"))
				}),
			},
			connectErr: false,
		},
		{
			name: "successful connection using ECDSA key",
			config: Config{
				Host:            "localhost",
				Port:            2222,
				Username:        "resticprofile",
				KnownHostsPath:  filepath.Join(tmpDir, "known_hosts"),
				PrivateKeyPaths: []string{filepath.Join(tmpDir, "id_ecdsa")},
				Handler: http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
					resp.Write([]byte("Connection successful using ECDSA key\n"))
				}),
			},
			connectErr: false,
		},
		{
			name: "successful connection using ED25519 key",
			config: Config{
				Host:            "localhost",
				Port:            2222,
				Username:        "resticprofile",
				KnownHostsPath:  filepath.Join(tmpDir, "known_hosts"),
				PrivateKeyPaths: []string{filepath.Join(tmpDir, "id_ed25519")},
				Handler: http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
					resp.Write([]byte("Connection successful using ED25519 key\n"))
				}),
			},
			connectErr: false,
		},
		{
			name: "successful connection using any of the provided key",
			config: Config{
				Host:           "localhost",
				Port:           2222,
				Username:       "resticprofile",
				KnownHostsPath: filepath.Join(tmpDir, "known_hosts"),
				PrivateKeyPaths: []string{
					filepath.Join(tmpDir, "file-not-found"), // Next key should be used
					filepath.Join(tmpDir, "id_ed25519"),
					filepath.Join(tmpDir, "id_ecdsa"),
					filepath.Join(tmpDir, "id_rsa"),
				},
				Handler: http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
					resp.Write([]byte("Connection successful any of the provided key\n"))
				}),
				ConnectTimeout: 10 * time.Second,
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

func TestSSHClientRunCommandWithCancelledContext(t *testing.T) {
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

	tmpDir := os.Getenv("SSH_TESTS_TMPDIR")
	if tmpDir == "" {
		tmpDir = filepath.Join(os.TempDir(), "resticprofile-ssh-tests")
	}

	config := Config{
		Host:           "localhost",
		Port:           2222,
		Username:       "resticprofile",
		KnownHostsPath: filepath.Join(tmpDir, "known_hosts"),
		PrivateKeyPaths: []string{
			filepath.Join(tmpDir, "id_ed25519"),
			filepath.Join(tmpDir, "id_ecdsa"),
			filepath.Join(tmpDir, "id_rsa"),
		},
		Handler: http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			t.Error("should not have been called")
		}),
	}

	for _, client := range []Client{NewOpenSSHClient(config), NewInternalClient(config)} {
		t.Run(client.Name(), func(t *testing.T) {
			defer client.Close(context.Background())

			ctx, cancel := context.WithCancel(context.Background())

			err := client.Connect(ctx)
			require.NoError(t, err)

			cancel()

			err = client.Run(ctx, "curl", fmt.Sprintf("http://localhost:%d/", client.TunnelPeerPort()))
			require.Error(t, err)
			assert.ErrorIs(t, err, context.Canceled)
		})
	}
}

func TestSSHClientRunCommandThenCancelContext(t *testing.T) {
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

	tmpDir := os.Getenv("SSH_TESTS_TMPDIR")
	if tmpDir == "" {
		tmpDir = filepath.Join(os.TempDir(), "resticprofile-ssh-tests")
	}

	config := Config{
		Host:           "localhost",
		Port:           2222,
		Username:       "resticprofile",
		KnownHostsPath: filepath.Join(tmpDir, "known_hosts"),
		PrivateKeyPaths: []string{
			filepath.Join(tmpDir, "id_ed25519"),
			filepath.Join(tmpDir, "id_ecdsa"),
			filepath.Join(tmpDir, "id_rsa"),
		},
	}

	for _, client := range []Client{NewOpenSSHClient(config), NewInternalClient(config)} {
		t.Run(client.Name(), func(t *testing.T) {
			defer client.Close(context.Background())

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			err := client.Connect(ctx)
			require.NoError(t, err)

			err = client.Run(ctx, "sleep", "10")
			require.Error(t, err)
			t.Log(err)
		})
	}
}
