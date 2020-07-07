package clog

import (
	"io"
	"log"
	"os"
)

// FileLog logs messages to a file
type FileLog struct {
	quiet   bool
	verbose bool
	file    *os.File
	writer  io.Writer
}

// NewFileLog creates a new file logger
//  Remember to Close() the logger at the end
func NewFileLog(filename string) (*FileLog, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	logger := &FileLog{
		file:   file,
		writer: file,
	}
	// also redirect the standard logger to the file
	SetOutput(file)
	return logger, nil
}

// NewStreamLog creates a new stream logger
func NewStreamLog(w io.Writer) *FileLog {
	logger := &FileLog{
		writer: w,
	}
	// also redirect the standard logger to the stream
	SetOutput(w)
	return logger
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

// Quiet will only display warnings and errors
func (l *FileLog) Quiet() {
	l.quiet = true
	l.verbose = false
}

// Verbose will display debugging information
func (l *FileLog) Verbose() {
	l.verbose = true
	l.quiet = false
}

// Debug sends debugging information
func (l *FileLog) Debug(v ...interface{}) {
	if !l.verbose {
		return
	}
	l.message(DebugLevel, v...)
}

// Debugf sends debugging information
func (l *FileLog) Debugf(format string, v ...interface{}) {
	if !l.verbose {
		return
	}
	l.messagef(DebugLevel, format, v...)
}

// Info logs some noticeable information
func (l *FileLog) Info(v ...interface{}) {
	if l.quiet {
		return
	}
	l.message(InfoLevel, v...)
}

// Infof logs some noticeable information
func (l *FileLog) Infof(format string, v ...interface{}) {
	if l.quiet {
		return
	}
	l.messagef(InfoLevel, format, v...)
}

// Warning send some important message to the console
func (l *FileLog) Warning(v ...interface{}) {
	l.message(WarningLevel, v...)
}

// Warningf send some important message to the console
func (l *FileLog) Warningf(format string, v ...interface{}) {
	l.messagef(WarningLevel, format, v...)
}

// Error sends error information to the console
func (l *FileLog) Error(v ...interface{}) {
	l.message(ErrorLevel, v...)
}

// Errorf sends error information to the console
func (l *FileLog) Errorf(format string, v ...interface{}) {
	l.messagef(ErrorLevel, format, v...)
}

func (l *FileLog) message(level int, v ...interface{}) {
	v = append([]interface{}{getLevelName(level)}, v...)
	log.Println(v...)
}

func (l *FileLog) messagef(level int, format string, v ...interface{}) {
	log.Printf(getLevelName(level)+" "+format+"\n", v...)
}

// Verify interface
var (
	_ Log = &FileLog{}
)
