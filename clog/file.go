package clog

import (
	"os"
)

// FileLog logs messages to a file
type FileLog struct {
	*StreamLog
	file *os.File
}

// NewFileLog creates a new file logger
//  Remember to Close() the logger at the end
func NewFileLog(filename string) (*FileLog, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	// standard output is managed by StreamLog
	logger := &FileLog{
		StreamLog: NewStreamLog(file),
	}
	return logger, nil
}

// Close the logfile when no longer needed
//  please note this method reinstate the NULL logger as the default logger
func (l *FileLog) Close() {
	if l.file != nil {
		l.file.Sync()
		l.file.Close()
		l.file = nil
	}
	l.writer = nil
	// make sure any other call to the logger won't panic
	defaultLog = &NullLog{}
	// and reset the standard logger
	SetOutput(os.Stdout)
}

// Verify interface
var (
	_ Logger = &FileLog{}
)
