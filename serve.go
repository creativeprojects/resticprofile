package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/creativeprojects/clog"
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
			resp.Write([]byte("remote not found"))
			return
		}
		remote, err := cmdCtx.config.GetRemote(remoteName)
		if err != nil {
			resp.Header().Set("Content-Type", "text/plain")
			resp.WriteHeader(http.StatusBadRequest)
			resp.Write([]byte(err.Error()))
			return
		}

		clog.Debugf("sending configuration for %q", remoteName)
		resp.Header().Set("Content-Type", "application/x-tar")
		resp.WriteHeader(http.StatusOK)
		sendFiles(resp, remote.SendFiles)
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	defer signal.Stop(quit)

	server := &http.Server{
		Addr:    fmt.Sprintf("localhost:%s", cmdCtx.flags.resticArgs[1]),
		Handler: handler,
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
	clog.Debugf("listening on %s", server.Addr)
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
