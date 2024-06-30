package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/creativeprojects/resticprofile/ssh"
)

func sendProfileCommand(w io.Writer, cmdCtx commandContext) error {
	if len(cmdCtx.flags.resticArgs) < 2 {
		return fmt.Errorf("missing argument: remote name")
	}
	if !cmdCtx.config.HasRemote(cmdCtx.flags.resticArgs[1]) {
		return fmt.Errorf("remote not found")
	}
	remote, err := cmdCtx.config.GetRemote(cmdCtx.flags.resticArgs[1])
	if err != nil {
		return err
	}
	// send the files to the remote using tar
	handler := http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "application/x-tar")
		resp.WriteHeader(http.StatusOK)
		sendFiles(resp, remote.SendFiles)
	})
	cnx := ssh.NewSSH(ssh.Config{
		Host:           remote.Host,
		Username:       remote.Username,
		PrivateKeyPath: remote.PrivateKeyPath,
		KnownHostsPath: remote.KnownHostsPath,
		Handler:        handler,
	})
	defer cnx.Close()

	err = cnx.Connect()
	if err != nil {
		return err
	}
	return nil
}
