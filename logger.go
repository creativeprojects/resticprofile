package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/dial"
	"github.com/creativeprojects/resticprofile/platform"
	"github.com/creativeprojects/resticprofile/remote"
	"github.com/creativeprojects/resticprofile/term"
	"github.com/creativeprojects/resticprofile/util"
	"github.com/creativeprojects/resticprofile/util/collect"
	"github.com/fatih/color"
)

type LogCloser interface {
	clog.Handler
	Close() error
}

func setupConsoleLogger(flags commandLineFlags) {
	if flags.stderr {
		out := color.Output
		color.Output = color.Error
		defer func() { color.Output = out }()
	}

	consoleHandler := clog.NewConsoleHandler("", log.LstdFlags)
	if flags.theme != "" {
		consoleHandler.SetTheme(flags.theme)
	}
	if flags.noAnsi {
		consoleHandler.Colouring(false)
	}
	logger := newFilteredLogger(flags, consoleHandler)
	clog.SetDefaultLogger(logger)
}

func setupRemoteLogger(flags commandLineFlags, client *remote.Client) {
	client.SetPrefix("elevated user: ")
	logger := newFilteredLogger(flags, clog.NewLogger(client))
	clog.SetDefaultLogger(logger)
}

func setupTargetLogger(flags commandLineFlags, logTarget, logUploadTarget, commandOutput string) (io.Closer, error) {
	var (
		handler  LogCloser
		file     io.Writer
		filepath string
		err      error
	)
	if scheme, hostPort, isURL := dial.GetAddr(logTarget); isURL {
		handler, file, err = getSyslogHandler(scheme, hostPort)
	} else if dial.IsURL(logTarget) {
		err = fmt.Errorf("unsupported URL: %s", logTarget)
	} else {
		filepath = getLogfilePath(logTarget)
		handler, file, err = getFileHandler(filepath)
	}
	if err != nil {
		return nil, err
	}
	// use the console handler as a backup
	logger := newFilteredLogger(flags, clog.NewSafeHandler(handler, clog.NewConsoleHandler("", log.LstdFlags)))
	// default logger added with level filtering
	clog.SetDefaultLogger(logger)

	// also redirect all terminal output
	if file != nil {
		if all, toLog := parseCommandOutput(commandOutput); all {
			term.SetOutput(io.MultiWriter(file, term.GetOutput()))
			term.SetErrorOutput(io.MultiWriter(file, term.GetErrorOutput()))
		} else if toLog {
			term.SetAllOutput(file)
		}
		if logUploadTarget != "" && filepath != "" {
			if !dial.IsURL(logUploadTarget) {
				return nil, fmt.Errorf("log-upload: No valid URL %v", logUploadTarget)
			}
			handler = createLogUploadingLogHandler(handler, filepath, logUploadTarget)
		}
	}
	// and return the handler (so we can close it at the end)
	return handler, nil
}

type logUploadingLogCloser struct {
	LogCloser
	logfilePath     string
	logUploadTarget string
}

// Try to close the original handler
// Also upload the log to the configured log-upload-url
func (w logUploadingLogCloser) Close() error {
	err := w.LogCloser.Close()
	if err != nil {
		return err
	}
	// Open logfile for reading
	logData, err := os.Open(w.logfilePath)
	if err != nil {
		return err
	}
	// Upload logfile to server
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(w.logUploadTarget, "application/octet-stream", logData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// HTTP-Status-Codes 200-299 signal success, return an error for everything else
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("log-upload: Got invalid http status %v: %v", resp.StatusCode, string(respBody))
	}
	return nil
}

func createLogUploadingLogHandler(handler LogCloser, logfilePath string, logUploadTarget string) LogCloser {
	return logUploadingLogCloser{LogCloser: handler, logfilePath: logfilePath, logUploadTarget: logUploadTarget}
}

func parseCommandOutput(commandOutput string) (all, log bool) {
	if strings.TrimSpace(commandOutput) == "auto" {
		if term.OsStdoutIsTerminal() {
			commandOutput = "log,console"
		} else {
			commandOutput = "log"
		}
	}
	co := collect.From(strings.Split(commandOutput, ","), strings.TrimSpace)
	log = slices.Contains(co, "log")
	all = slices.Contains(co, "all") || (log && slices.Contains(co, "console"))
	return
}

func getLogfilePath(logfile string) string {
	if strings.HasPrefix(logfile, constants.TemporaryDirMarker) {
		if tempDir, err := util.TempDir(); err == nil {
			logfile = logfile[len(constants.TemporaryDirMarker):]
			if len(logfile) > 0 && os.IsPathSeparator(logfile[0]) {
				logfile = logfile[1:]
			}
			logfile = filepath.Join(tempDir, logfile)
			_ = os.MkdirAll(filepath.Dir(logfile), 0755)
		}
	}
	return logfile
}

func getFileHandler(logfile string) (*clog.StandardLogHandler, io.Writer, error) {
	// create a platform aware log file appender
	keepOpen, appender := true, appendFunc(nil)
	if platform.IsWindows() {
		keepOpen = false
		appender = func(dst []byte, c byte) []byte {
			switch c {
			case '\n':
				return append(dst, '\r', '\n') // normalize to CRLF on Windows
			case '\r':
				return dst
			}
			return append(dst, c)
		}
	}

	writer, err := newDeferredFileWriter(logfile, keepOpen, appender)
	if err != nil {
		return nil, nil, err
	}

	return clog.NewStandardLogHandler(writer, "", log.LstdFlags), writer, nil
}

func newFilteredLogger(flags commandLineFlags, handler clog.Handler) *clog.Logger {
	if flags.quiet && (flags.verbose || flags.veryVerbose) {
		var coin string
		if randomBool() {
			coin = "verbose"
			flags.quiet = false
		} else {
			coin = "quiet"
			flags.verbose = false
			flags.veryVerbose = false
		}
		// the logger hasn't been created yet, so we call the handler directly
		_ = handler.LogEntry(clog.LogEntry{
			Level:  clog.LevelWarning,
			Format: "you specified -quiet (-q) and -verbose (-v) at the same time. So let's flip a coin! ... and the winner is ... %s.",
			Values: []interface{}{coin},
		})
	}
	minLevel := clog.LevelInfo
	if flags.quiet {
		minLevel = clog.LevelWarning
	} else if flags.veryVerbose {
		minLevel = clog.LevelTrace
	} else if flags.verbose {
		minLevel = clog.LevelDebug
	}
	// now create and return the logger
	return clog.NewLogger(clog.NewLevelFilter(minLevel, handler))
}

func changeLevelFilter(level clog.LogLevel) {
	handler := clog.GetDefaultLogger().GetHandler()
	filter, ok := handler.(*clog.LevelFilter)
	if ok {
		filter.SetLevel(level)
	}
}

// deferredFileWriter accumulates Write requests and writes them at a fixed rate (every 250 ms)
type deferredFileWriter struct {
	done, flush chan chan error
	data        chan []byte
}

func (d *deferredFileWriter) Close() error {
	req := make(chan error)
	d.done <- req
	return <-req
}

func (d *deferredFileWriter) Flush() error {
	req := make(chan error)
	d.flush <- req
	return <-req
}

func (d *deferredFileWriter) Write(data []byte) (n int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	c := make([]byte, len(data))
	n = copy(c, data)
	d.data <- c
	return
}

type appendFunc func(dst []byte, c byte) []byte

func newDeferredFileWriter(filename string, keepOpen bool, appender appendFunc) (io.WriteCloser, error) {
	d := &deferredFileWriter{
		flush: make(chan chan error),
		done:  make(chan chan error),
		data:  make(chan []byte, 64),
	}

	var (
		buffer    []byte
		lastError error
		file      *os.File
	)

	closeFile := func() {
		if file != nil {
			lastError = file.Close()
			file = nil
		}
	}

	flush := func(alsoEmpty bool) {
		if len(buffer) == 0 && !alsoEmpty {
			return
		}
		if file == nil {
			file, lastError = os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		}
		if file != nil {
			var written int
			written, lastError = file.Write(buffer)
			if written == len(buffer) {
				buffer = buffer[:0]
			} else {
				buffer = buffer[written:]
			}
		}
		if keepOpen {
			_ = file.Sync()
		} else {
			closeFile()
		}
	}

	// test if we can create the file
	buffer = make([]byte, 0, 4096)
	flush(true)

	// data appending
	addToBuffer := func(data []byte) {
		buffer = append(buffer, data...) // fast path
	}
	if appender != nil {
		addToBuffer = func(data []byte) {
			for _, c := range data {
				buffer = appender(buffer, c)
			}
		}
	}

	addPendingData := func(size int) {
		for ; size > 0; size-- {
			select {
			case data, ok := <-d.data:
				if ok {
					addToBuffer(data)
				} else {
					return // closed
				}
			default:
				return // no-more-data
			}
		}
	}

	// data transport
	go func() {
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case data := <-d.data:
				addToBuffer(data)
			case <-ticker.C:
				flush(false)
			case req := <-d.flush:
				addPendingData(1024)
				flush(false)
				req <- lastError
			case req := <-d.done:
				close(d.done)
				close(d.flush)
				close(d.data)
				addPendingData(1024)
				flush(false)
				closeFile()
				req <- lastError
				return
			}
		}
	}()

	return d, lastError
}
