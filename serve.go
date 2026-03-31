package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/remote"
	"github.com/creativeprojects/resticprofile/ssh"
)

func serveCommand(w io.Writer, cmdCtx commandContext) error {
	if len(cmdCtx.flags.resticArgs) < 2 {
		return fmt.Errorf("missing argument: port")
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	defer signal.Stop(quit)

	return serveProfiles(cmdCtx.flags.resticArgs[1], cmdCtx.config, quit)
}

func serveProfiles(port string, config *config.Config, quit chan os.Signal) error {
	handler := http.NewServeMux()
	handler.HandleFunc("GET /configuration/{remote}", func(resp http.ResponseWriter, req *http.Request) {
		remoteName := req.PathValue("remote")
		if !config.HasRemote(remoteName) {
			sendError(resp, http.StatusNotFound, fmt.Errorf("remote %q not found", remoteName))
			return
		}
		remoteConfig, err := config.GetRemote(remoteName)
		if err != nil {
			sendError(resp, http.StatusBadRequest, fmt.Errorf("error while getting remote configuration: %w", err))
			return
		}

		sendRemoteFiles(remoteConfig, remoteName, nil, resp)
	})

	server := &http.Server{
		Addr:              "localhost:" + port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// put the shutdown code in a goroutine
	go func(server *http.Server, quit chan os.Signal) {
		<-quit

		clog.Info("shutting down the server")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := server.Shutdown(ctx)
		if err != nil {
			clog.Errorf("error while shutting down the server: %v", err)
		}

	}(server, quit)

	// we want to return the server error if any so we need to keep it in the main thread.
	clog.Infof("listening on %s", server.Addr)
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func sendProfileCommand(w io.Writer, cmdCtx commandContext) error {
	if len(cmdCtx.flags.resticArgs) < 2 {
		return fmt.Errorf("missing argument: remote name")
	}
	remoteName := cmdCtx.flags.resticArgs[1]
	if !cmdCtx.config.HasRemote(remoteName) {
		return fmt.Errorf("remote not found")
	}
	remoteConfig, err := cmdCtx.config.GetRemote(remoteName)
	if err != nil {
		return err
	}
	// send the files to the remote using tar
	handler := http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		sendRemoteFiles(remoteConfig, remoteName, cmdCtx.flags.resticArgs[2:], resp)
	})
	sshConfig := ssh.Config{
		Host:            remoteConfig.Host,
		Port:            remoteConfig.Port,
		Username:        remoteConfig.Username,
		PrivateKeyPaths: remoteConfig.PrivateKeyPaths,
		KnownHostsPath:  remoteConfig.KnownHostsPath,
		SSHConfigPath:   remoteConfig.SSHConfig,
		Handler:         handler,
	}
	var cnx ssh.Client
	switch remoteConfig.Connection {
	case "ssh":
		cnx = ssh.NewInternalClient(sshConfig)
	case "openssh":
		cnx = ssh.NewOpenSSHClient(sshConfig)
	default:
		return fmt.Errorf("unsupported connection type %q for remote %q", remoteConfig.Connection, remoteName)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGABRT)
	defer cancel()

	err = cnx.Connect(ctx)
	defer cnx.Close(context.WithoutCancel(ctx))
	if err != nil {
		return err
	}

	binaryPath := remoteConfig.BinaryPath
	if binaryPath == "" {
		binaryPath = "resticprofile"
	}
	arguments := []string{
		"-v",
		"-r", fmt.Sprintf("http://localhost:%d/configuration/%s", cnx.TunnelPeerPort(), remoteName),
	}
	err = cnx.Run(ctx, binaryPath, arguments...)
	if err != nil {
		return fmt.Errorf("failed to run resticprofile on peer: %w", err)
	}
	return nil
}

func sendRemoteFiles(remoteConfig *config.Remote, remoteName string, extraArgs []string, resp http.ResponseWriter) {
	// prepare manifest file
	manifest := remote.Manifest{
		Version:              version,
		ConfigurationFile:    path.Base(remoteConfig.ConfigurationFile), // need to take file path into consideration
		ProfileName:          remoteConfig.ProfileName,
		CommandLineArguments: extraArgs,
	}
	manifestData, err := json.Marshal(manifest)
	if err != nil {
		sendError(resp, http.StatusInternalServerError, fmt.Errorf("error while generating manifest: %w", err))
		return
	}

	clog.Debugf("sending configuration for %q", remoteName)

	tar := remote.NewTar(resp)
	err = tar.PrepareFiles(append(remoteConfig.SendFiles, remoteConfig.ConfigurationFile))
	if err != nil {
		sendError(resp, http.StatusInternalServerError, fmt.Errorf("error while preparing files to send for remote %q: %w", remoteName, err))
		return
	}
	defer tar.Close()

	resp.Header().Set("Content-Type", "application/x-tar")
	resp.WriteHeader(http.StatusOK)

	err = tar.SendFiles()
	if err != nil {
		clog.Error(err)
		return
	}
	err = tar.SendFile(constants.ManifestFilename, manifestData)
	if err != nil {
		clog.Error(err)
		return
	}
}

func sendError(resp http.ResponseWriter, status int, err error) {
	resp.Header().Set("Content-Type", "text/plain")
	resp.WriteHeader(status)
	_, _ = resp.Write([]byte(err.Error())) //nolint:gosec // G705: XSS via taint analysis
	_, _ = resp.Write([]byte("\n"))
	clog.Error(err)
}
