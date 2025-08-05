package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/remote"
	"github.com/creativeprojects/resticprofile/ssh"
)

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
		// prepare manifest file
		manifest := remote.Manifest{
			ConfigurationFile:    path.Base(remoteConfig.ConfigurationFile), // need to take file path into consideration
			ProfileName:          remoteConfig.ProfileName,
			CommandLineArguments: cmdCtx.flags.resticArgs[2:],
		}
		manifestData, err := json.Marshal(manifest)
		if err != nil {
			resp.Header().Set("Content-Type", "text/plain")
			resp.WriteHeader(http.StatusInternalServerError)
			_, _ = resp.Write([]byte(err.Error()))
			return
		}

		resp.Header().Set("Content-Type", "application/x-tar")
		resp.WriteHeader(http.StatusOK)

		tar := remote.NewTar(resp)
		defer tar.Close()
		_ = tar.SendFiles(append(remoteConfig.SendFiles, remoteConfig.ConfigurationFile))
		_ = tar.SendFile(constants.ManifestFilename, manifestData)
	})
	cnx := ssh.NewSSH(ssh.Config{
		Host:           remoteConfig.Host,
		Username:       remoteConfig.Username,
		PrivateKeyPath: remoteConfig.PrivateKeyPath,
		KnownHostsPath: remoteConfig.KnownHostsPath,
		Handler:        handler,
	})
	defer cnx.Close()

	err = cnx.Connect()
	if err != nil {
		return err
	}
	binaryPath := remoteConfig.BinaryPath
	if binaryPath == "" {
		binaryPath = "resticprofile"
	}
	commandLine := fmt.Sprintf("%s -v -r http://localhost:%d/configuration/%s ",
		binaryPath,
		cnx.TunnelPort(),
		remoteName,
	)
	err = cnx.Run(commandLine)
	if err != nil {
		return fmt.Errorf("failed to run command %q: %w", commandLine, err)
	}
	return nil
}
