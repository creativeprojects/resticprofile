package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/term"
)

func (r *resticWrapper) prepareStreamSource() (io.ReadCloser, error) {
	if r.profile.Backup != nil && r.profile.Backup.UseStdin {
		if len(r.profile.Backup.StdinCommand) > 0 {
			return r.prepareCommandStreamSource()
		} else {
			return r.prepareStdinStreamSource()
		}
	}
	return nil, nil
}

func (r *resticWrapper) prepareStdinStreamSource() (io.ReadCloser, error) {
	clog.Debug("redirecting stdin to the backup")

	if r.stdin == nil {
		return nil, fmt.Errorf("stdin was already consumed. cannot read it twice")
	}

	totalBytes := int64(0)

	readCloser := &readerWrapper{
		reader: r.stdin,
		read: func(w *readerWrapper, bytes []byte) (n int, err error) {
			n, err = w.reader.Read(bytes)
			totalBytes += int64(n)
			return
		},
		close: func(w *readerWrapper) error {
			if totalBytes > 0 && r.stdin != nil {
				r.stdin = nil
				w.close = nil
			}
			return nil
		},
	}

	return readCloser, nil
}

func (r *resticWrapper) prepareCommandStreamSource() (io.ReadCloser, error) {
	clog.Debug("redirecting command output to the backup")
	pipeReader, pipeWriter := io.Pipe()
	bufferedWriter := bufio.NewWriterSize(pipeWriter, 8*1024)

	commandSignals := make(chan os.Signal, 2)
	signal.Notify(commandSignals, os.Interrupt, syscall.SIGTERM, syscall.SIGABRT)

	go func() {
		defer pipeWriter.Close()
		defer bufferedWriter.Flush()
		defer signal.Stop(commandSignals)

		env := append(os.Environ(), r.getEnvironment()...)
		env = append(env, r.getProfileEnvironment()...)

		for i, sourceCommand := range r.profile.Backup.StdinCommand {
			clog.Debugf("starting 'stdin-command' command %d/%d: %s", i+1, len(r.profile.Backup.StdinCommand), sourceCommand)
			rCommand := newShellCommand(sourceCommand, nil, env, r.dryRun, commandSignals, nil)
			rCommand.stdout = bufferedWriter
			rCommand.stderr = term.GetErrorOutput()

			_, stderr, err := runShellCommand(rCommand)
			if err != nil {
				// discard unflushed output
				bufferedWriter.Reset(pipeWriter)
				// push command error to reader
				err = newCommandError(rCommand, stderr, fmt.Errorf("'stdin-command' on profile '%s': %w", r.profile.Name, err))
				if closeError := pipeWriter.CloseWithError(err); closeError != nil {
					clog.Errorf("Failed closing pipe for command '%s' after %w ; close error: %w", sourceCommand, err, closeError)
				}
				return
			}
		}
	}()

	closePipe := func() error {
		defer func() {
			clog.Debugf("stopping 'stdin-command'")
			signal.Stop(commandSignals)
			commandSignals <- os.Interrupt
		}()
		return pipeReader.Close()
	}

	// read from pipe to ensure the process started and returns content or error before restic is started
	var initialReader io.Reader
	{
		initialBytes := make([]byte, 512)
		if n, err := pipeReader.Read(initialBytes); err == nil || err == io.EOF {
			clog.Debugf("initial %d bytes successfully read from 'stdin-command'", n)
			initialReader = bytes.NewReader(initialBytes[:n])
		} else {
			_ = closePipe()
			return nil, err
		}
	}

	readCloser := &readerWrapper{
		reader: io.MultiReader(initialReader, pipeReader),

		read: func(w *readerWrapper, bytes []byte) (n int, err error) {
			n, err = w.reader.Read(bytes)

			// Stopping restic when stream source command fails while producing content
			if err != nil && err != io.EOF {
				clog.Errorf("interrupting '%s' after stdin read error: %s", r.command, err)
				if runtime.GOOS == "windows" {
					return // that will close stdin and stops restic
				} else if r.sigChan != nil {
					r.sigChan <- os.Interrupt
				}
				// Wait for the signal to arrive before allowing further read from stdin
				time.Sleep(750 * time.Millisecond)
			}
			return
		},

		close: func(w *readerWrapper) error {
			w.close = nil
			return closePipe()
		},
	}

	return readCloser, nil
}

type readerWrapper struct {
	reader              io.Reader
	readLock, closeLock sync.Mutex
	read                func(w *readerWrapper, bytes []byte) (n int, err error)
	close               func(w *readerWrapper) error
}

func (w *readerWrapper) Read(bytes []byte) (n int, err error) {
	w.readLock.Lock()
	defer w.readLock.Unlock()

	return w.read(w, bytes)
}

func (w *readerWrapper) Close() error {
	w.closeLock.Lock()
	defer w.closeLock.Unlock()

	if w.close != nil {
		return w.close(w)
	}
	return nil
}
