package clog

type NullLog struct{}

// Quiet will only display warnings and errors
func (l *NullLog) Quiet() {
}

// Verbose will display debugging information
func (l *NullLog) Verbose() {
}

// Debug sends debugging information
func (l *NullLog) Debug(v ...interface{}) {
}

// Debugf sends debugging information
func (l *NullLog) Debugf(format string, v ...interface{}) {
}

// Info logs some noticeable information
func (l *NullLog) Info(v ...interface{}) {
}

// Infof logs some noticeable information
func (l *NullLog) Infof(format string, v ...interface{}) {
}

// Warning send some important message to the console
func (l *NullLog) Warning(v ...interface{}) {
}

// Warningf send some important message to the console
func (l *NullLog) Warningf(format string, v ...interface{}) {
}

// Error sends error information to the console
func (l *NullLog) Error(v ...interface{}) {
}

// Errorf sends error information to the console
func (l *NullLog) Errorf(format string, v ...interface{}) {
}

// Verify interface
var (
	_ Log = &NullLog{}
)
