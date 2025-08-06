package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"sync"
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
	handler := http.NewServeMux()
	handler.HandleFunc("GET /configuration/{remote}", func(resp http.ResponseWriter, req *http.Request) {
		remoteName := req.PathValue("remote")
		if !cmdCtx.config.HasRemote(remoteName) {
			sendError(resp, http.StatusNotFound, fmt.Errorf("remote %q not found", remoteName))
			return
		}
		remoteConfig, err := cmdCtx.config.GetRemote(remoteName)
		if err != nil {
			sendError(resp, http.StatusBadRequest, fmt.Errorf("error while getting remote configuration: %w", err))
			return
		}

		sendRemoteFiles(remoteConfig, remoteName, nil, resp)
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	defer signal.Stop(quit)

	server := &http.Server{
		Addr:              fmt.Sprintf("localhost:%s", cmdCtx.flags.resticArgs[1]),
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
		Host:           remoteConfig.Host,
		Username:       remoteConfig.Username,
		PrivateKeyPath: remoteConfig.PrivateKeyPath,
		KnownHostsPath: remoteConfig.KnownHostsPath,
		SSHConfigPath:  remoteConfig.SSHConfig,
		Handler:        handler,
	}
	// cnx := ssh.NewSSH(sshConfig)
	// defer cnx.Close()

	// err = cnx.Connect()
	// if err != nil {
	// 	return err
	// }
	// binaryPath := remoteConfig.BinaryPath
	// if binaryPath == "" {
	// 	binaryPath = "resticprofile"
	// }
	// commandLine := fmt.Sprintf("%s -v -r http://localhost:%d/configuration/%s ",
	// 	binaryPath,
	// 	cnx.TunnelPort(),
	// 	remoteName,
	// )
	// err = cnx.Run(commandLine)
	// if err != nil {
	// 	return fmt.Errorf("failed to run command %q: %w", commandLine, err)
	// }
	err = callOpenSSH(context.Background(), sshConfig, handler)
	if err != nil {
		return err
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
	resp.Header().Set("Content-Type", "application/x-tar")
	resp.WriteHeader(http.StatusOK)

	tar := remote.NewTar(resp)
	defer tar.Close()
	_ = tar.SendFiles(append(remoteConfig.SendFiles, remoteConfig.ConfigurationFile))
	_ = tar.SendFile(constants.ManifestFilename, manifestData)
}

func sendError(resp http.ResponseWriter, status int, err error) {
	resp.Header().Set("Content-Type", "text/plain")
	resp.WriteHeader(status)
	_, _ = resp.Write([]byte(err.Error()))
	_, _ = resp.Write([]byte("\n"))
	clog.Error(err)
}

func callOpenSSH(ctx context.Context, config ssh.Config, handler http.HandlerFunc) error {
	var wg sync.WaitGroup
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return err
	}
	server := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer ln.Close()
		err := server.Serve(ln)
		if err != nil && err != http.ErrServerClosed {
			clog.Error("error while serving HTTP:", err)
		}
	}()
	ctx = context.WithoutCancel(ctx)
	err = server.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("error while shutting down server: %w", err)
	}
	cmd := exec.CommandContext(ctx,
		"ssh", "-v",
		"-F", config.SSHConfigPath,
		"-R", fmt.Sprintf("0:localhost:%d", ln.Addr().(*net.TCPAddr).Port),
		fmt.Sprintf("%s@%s", config.Username, config.Host),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	clog.Debugf("running command: %s", cmd.String())
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error while running ssh command: %w", err)
	}
	wg.Wait()
	return nil
}
