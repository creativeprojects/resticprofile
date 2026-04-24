package write

import (
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/creativeprojects/resticprofile/platform"
)

var ErrAttemptToWriteOnClosedFile = errors.New("cannot write to a closed or unopened file")

type File struct {
	filename        string
	perm            os.FileMode
	flag            int
	keepOpen        bool
	keepOpenTimeout time.Duration
	handle          *os.File
	mutex           sync.Mutex
	timer           *time.Timer
	timerMutex      sync.Mutex
	// stats
	fileOpenCount  atomic.Int32
	fileCloseCount atomic.Int32
}

func NewFile(filename string, options ...FileOption) (f *File, err error) {
	f = &File{
		filename:        filename,
		perm:            0644,
		flag:            os.O_WRONLY | os.O_APPEND | os.O_CREATE,
		keepOpen:        !platform.IsWindows(),
		keepOpenTimeout: 10 * time.Millisecond,
	}

	for _, option := range options {
		option(f)
	}

	err = f.open()
	if !f.keepOpen {
		defer func() {
			err = errors.Join(err, f.Close())
		}()
	}
	return
}

func (f *File) open() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.handle != nil {
		return nil
	}
	var err error
	f.fileOpenCount.Add(1)
	f.handle, err = os.OpenFile(f.filename, f.flag, f.perm)
	return err
}

func (f *File) Close() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.handle != nil {
		f.fileCloseCount.Add(1)
		err := f.handle.Close()
		f.handle = nil
		return err
	}
	return nil
}

func (f *File) Flush() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.handle != nil {
		return f.handle.Sync()
	}
	return nil
}

func (f *File) Write(data []byte) (n int, err error) {
	if !f.keepOpen {
		f.stopCloseTimer()
		err := f.open()
		if err != nil {
			return 0, err
		}
		defer f.resetCloseTimer()
	}

	if f.handle == nil {
		return 0, ErrAttemptToWriteOnClosedFile
	}

	n, err = f.handle.Write(data)
	return
}

func (f *File) stopCloseTimer() {
	f.timerMutex.Lock()
	defer f.timerMutex.Unlock()
	if f.timer != nil {
		f.timer.Stop()
	}
}

func (f *File) resetCloseTimer() {
	f.timerMutex.Lock()
	defer f.timerMutex.Unlock()

	if f.timer != nil {
		f.timer.Stop()
	}
	f.timer = time.AfterFunc(f.keepOpenTimeout, func() {
		_ = f.Close()
	})
}

// stats returns the number of times the file was opened and closed
func (f *File) stats() (int32, int32) {
	return f.fileOpenCount.Load(), f.fileCloseCount.Load()
}
