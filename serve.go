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
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/remote"
	"github.com/spf13/afero"
)

func serveCommand(w io.Writer, cmdCtx commandContext) error {
	if len(cmdCtx.flags.resticArgs) < 2 {
		return fmt.Errorf("missing argument: port")
	}
	handler := http.NewServeMux()
	handler.HandleFunc("GET /configuration/{remote}", func(resp http.ResponseWriter, req *http.Request) {
		remoteName := req.PathValue("remote")
		if !cmdCtx.config.HasRemote(remoteName) {
			resp.Header().Set("Content-Type", "text/plain")
			resp.WriteHeader(http.StatusNotFound)
			_, _ = resp.Write([]byte("remote not found"))
			return
		}
		remoteConfig, err := cmdCtx.config.GetRemote(remoteName)
		if err != nil {
			resp.Header().Set("Content-Type", "text/plain")
			resp.WriteHeader(http.StatusBadRequest)
			_, _ = resp.Write([]byte(err.Error()))
			return
		}

		// prepare manifest file
		manifest := remote.Manifest{
			ConfigurationFile: path.Base(remoteConfig.ConfigurationFile), // need to take file path into consideration
			ProfileName:       remoteConfig.ProfileName,
		}
		manifestData, err := json.Marshal(manifest)
		if err != nil {
			resp.Header().Set("Content-Type", "text/plain")
			resp.WriteHeader(http.StatusInternalServerError)
			_, _ = resp.Write([]byte(err.Error()))
			return
		}

		clog.Debugf("sending configuration for %q", remoteName)
		resp.Header().Set("Content-Type", "application/x-tar")
		resp.WriteHeader(http.StatusOK)

		tar := remote.NewTar(resp)
		defer tar.Close()
		_ = tar.SendFiles(afero.NewOsFs(), append(remoteConfig.SendFiles, remoteConfig.ConfigurationFile))
		_ = tar.SendFile(constants.ManifestFilename, manifestData)

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
