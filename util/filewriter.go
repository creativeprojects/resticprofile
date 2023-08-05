package util

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/creativeprojects/resticprofile/platform"
)

// AsyncFileWriterAppendFunc is called for every input byte when appending it to the output buffer (is similar to buf = append(buf, byte))
type AsyncFileWriterAppendFunc func(dst []byte, c byte) []byte

type asyncFileWriterOption func(writer *asyncFileWriter)

// WithAsyncWriteInterval sets the interval at which writes happen at least when data is pending
func WithAsyncWriteInterval(duration time.Duration) asyncFileWriterOption {
	return func(writer *asyncFileWriter) { writer.interval = duration }
}

// WithAsyncFileKeepOpen toggles whether the file is kept open between writes. Defaults to true for all OS except Windows.
func WithAsyncFileKeepOpen(keepOpen bool) asyncFileWriterOption {
	return func(writer *asyncFileWriter) { writer.keepOpen = keepOpen }
}

// WithAsyncFileAppendFunc sets AsyncFileWriterAppendFunc. Default is to not use a custom appender.
func WithAsyncFileAppendFunc(appender AsyncFileWriterAppendFunc) asyncFileWriterOption {
	return func(writer *asyncFileWriter) { writer.appender = appender }
}

// WithAsyncFilePerm sets file perms to apply when creating the file
func WithAsyncFilePerm(perm os.FileMode) asyncFileWriterOption {
	return func(writer *asyncFileWriter) { writer.perm = perm }
}

// WithAsyncFileFlag sets file open flags
func WithAsyncFileFlag(flag int) asyncFileWriterOption {
	return func(writer *asyncFileWriter) { writer.flag = flag }
}

// WithAsyncFileTruncate enables that existing files are truncated
func WithAsyncFileTruncate() asyncFileWriterOption {
	return func(writer *asyncFileWriter) { writer.flag |= os.O_TRUNC }
}

const (
	asyncWriterDataChanSize = 64
	asyncWriterBlockSize    = 4 * 1024
	asyncWriterMaxBlockSize = asyncWriterDataChanSize * asyncWriterBlockSize
)

var (
	asyncWriterBufferPool = sync.Pool{
		New: func() any { return make([]byte, asyncWriterBlockSize) },
	}
)

func asyncWriterReturnToPool(data []byte) {
	if cap(data) == asyncWriterBlockSize {
		asyncWriterBufferPool.Put(data[:0])
	}
}

// NewAsyncFileWriter creates a file writer that accumulates Write requests and writes them at a fixed rate (every 250 ms by default)
func NewAsyncFileWriter(filename string, options ...asyncFileWriterOption) (io.WriteCloser, error) {
	w := &asyncFileWriter{
		flush:    make(chan chan error),
		done:     make(chan chan error),
		data:     make(chan []byte, asyncWriterDataChanSize),
		perm:     0644,
		flag:     os.O_WRONLY | os.O_APPEND | os.O_CREATE,
		interval: 250 * time.Millisecond,
		keepOpen: !platform.IsWindows(),
	}
	for _, option := range options {
		option(w)
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

	flush := func(alsoEmpty, whenTooBig bool) {
		if len(buffer) == 0 && !alsoEmpty {
			return
		}
		if len(buffer) < asyncWriterMaxBlockSize && whenTooBig {
			return
		}
		if file == nil {
			file, lastError = os.OpenFile(filename, w.flag, w.perm)
		}
		if file != nil {
			var written int
			written, lastError = file.Write(buffer)
			if remaining := len(buffer) - written; remaining > 0 {
				copy(buffer, buffer[written:])
				buffer = buffer[:remaining]
			} else {
				buffer = buffer[:0]
			}
		}
		if w.keepOpen {
			_ = file.Sync()
		} else {
			closeFile()
		}
	}

	// test if we can create the file
	buffer = make([]byte, 0, asyncWriterBlockSize)
	flush(true, false)

	// data appending
	addToBuffer := func(data []byte) {
		buffer = append(buffer, data...) // fast path
		asyncWriterReturnToPool(data)
		flush(false, true)
	}
	if w.appender != nil {
		addToBuffer = func(data []byte) {
			for _, c := range data {
				buffer = w.appender(buffer, c)
			}
			asyncWriterReturnToPool(data)
			flush(false, true)
		}
	}

	addPendingData := func(max int) {
		for ; max > 0; max-- {
			select {
			case data, ok := <-w.data:
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
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()
		defer closeFile()

		for {
			select {
			case data := <-w.data:
				addToBuffer(data)
			case <-ticker.C:
				flush(false, false)
			case req := <-w.flush:
				addPendingData(1024)
				flush(false, false)
				req <- lastError
			case req := <-w.done:
				defer func(response chan error) {
					response <- lastError
				}(req)
				close(w.done)
				close(w.flush)
				close(w.data)
				addPendingData(1024)
				flush(false, false)
				closeFile()
				return
			}
		}
	}()

	return w, lastError
}

type asyncFileWriter struct {
	done, flush chan chan error
	data        chan []byte
	appender    AsyncFileWriterAppendFunc
	flag        int
	perm        os.FileMode
	keepOpen    bool
	interval    time.Duration
}

func (w *asyncFileWriter) Close() error {
	req := make(chan error)
	w.done <- req
	return <-req
}

func (w *asyncFileWriter) Flush() error {
	req := make(chan error)
	w.flush <- req
	return <-req
}

func (w *asyncFileWriter) Write(data []byte) (n int, _ error) {
	var buffer []byte
	if len(data) <= asyncWriterBlockSize {
		buffer = asyncWriterBufferPool.Get().([]byte)[:len(data)]
	} else {
		buffer = make([]byte, len(data))
	}

	n = copy(buffer, data)
	w.data <- buffer
	return
}
