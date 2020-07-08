package clog

// TestLog redirects all the logs to the testing framework logger
type TestLog struct {
	t TestLogInterface
}

// TestLogInterface for use with testing.B or testing.T
type TestLogInterface interface {
	Log(args ...interface{})
	Logf(format string, args ...interface{})
}

// NewTestLog instantiates a new logger redirecting to the test framework logger
// or any other implementation of TestLogInterface for that matter
func NewTestLog(t TestLogInterface) *TestLog {
	return &TestLog{
		t: t,
	}
}

// SetTestLog install a test logger as the default logger.
//  IMPORTANT: don't forget to run ClearTestLog() at the end of the test
func SetTestLog(t TestLogInterface) {
	defaultLog = NewTestLog(t)
}

// ClearTestLog at the end of the test otherwise the logger will keep a reference on t
func ClearTestLog() {
	defaultLog = &NullLog{}
}

// Log sends a log entry with the specified level
func (l *TestLog) Log(level LogLevel, v ...interface{}) {
	l.message(level, v...)
}

// Logf sends a log entry with the specified level
func (l *TestLog) Logf(level LogLevel, format string, v ...interface{}) {
	l.messagef(level, format, v...)
}

// Debug sends debugging information
func (l *TestLog) Debug(v ...interface{}) {
	l.message(DebugLevel, v...)
}

// Debugf sends debugging information
func (l *TestLog) Debugf(format string, v ...interface{}) {
	l.messagef(DebugLevel, format, v...)
}

// Info logs some noticeable information
func (l *TestLog) Info(v ...interface{}) {
	l.message(InfoLevel, v...)
}

// Infof logs some noticeable information
func (l *TestLog) Infof(format string, v ...interface{}) {
	l.messagef(InfoLevel, format, v...)
}

// Warning send some important message to the console
func (l *TestLog) Warning(v ...interface{}) {
	l.message(WarningLevel, v...)
}

// Warningf send some important message to the console
func (l *TestLog) Warningf(format string, v ...interface{}) {
	l.messagef(WarningLevel, format, v...)
}

// Error sends error information to the console
func (l *TestLog) Error(v ...interface{}) {
	l.message(ErrorLevel, v...)
}

// Errorf sends error information to the console
func (l *TestLog) Errorf(format string, v ...interface{}) {
	l.messagef(ErrorLevel, format, v...)
}

func (l *TestLog) message(level LogLevel, v ...interface{}) {
	v = append([]interface{}{getLevelName(level)}, v...)
	l.t.Log(v...)
}

func (l *TestLog) messagef(level LogLevel, format string, v ...interface{}) {
	l.t.Logf(getLevelName(level)+" "+format, v...)
}

// Verify interface
var (
	_ Logger = &TestLog{}
)
