package remote

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/creativeprojects/clog"
)

const (
	donePath = "/done"
	logPath  = "/log"
	termPath = "/term"
)

type logMessage struct {
	Level   int    `json:"level"`
	Message string `json:"message"`
}

var (
	serveMux *http.ServeMux
)

func getServeMux() *http.ServeMux {
	serveMux = http.NewServeMux()
	serveMux.HandleFunc(donePath, handlerFuncDone)
	serveMux.HandleFunc(logPath, handlerFuncLog)
	serveMux.HandleFunc(termPath, handlerFuncTerm)
	return serveMux
}

func handlerFuncDone(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Just close the http server
	StopServer()
}

func handlerFuncLog(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	log := &logMessage{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(log)
	if err != nil {
		clog.Errorf("error decoding json log message: %v", err)
	}
	switch clog.LogLevel(log.Level) {
	case clog.LevelTrace:
		clog.Trace(log.Message)

	case clog.LevelDebug:
		clog.Debug(log.Message)

	case clog.LevelInfo:
		clog.Info(log.Message)

	case clog.LevelWarning:
		clog.Warning(log.Message)

	case clog.LevelError:
		clog.Error(log.Message)

	default:
		clog.Log(clog.LevelInfo, log.Message)
	}
}

func handlerFuncTerm(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	_, err := io.Copy(os.Stdout, r.Body)
	if err != nil {
		clog.Errorf("error while copying terminal data: %w", err)
	}
}
