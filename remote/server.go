package remote

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/creativeprojects/clog"
)

const (
	timeout = 5
)

var (
	listener *net.TCPListener
	port     int
	server   *http.Server
)

// StartServer starts a http server
func StartServer(done chan interface{}) error {
	// let the system chose a port
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	listener, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	port = listener.Addr().(*net.TCPAddr).Port
	clog.Debugf("listening on port %d", port)

	server = &http.Server{
		Handler: getServeMux(),
	}
	go func() {
		_ = server.Serve(listener)
		close(done)
	}()

	return nil
}

// GetPort returns the port chosen by the system
func GetPort() int {
	return port
}

// StopServer gracefully asks the http server to shutdown
func StopServer() {
	if server != nil {
		// gracefully stop the http server
		ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}
	server = nil
	if listener != nil {
		listener.Close()
	}
	listener = nil
}
