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
	l.message(LevelDebug, v...)
}

// Debugf sends debugging information
func (l *TestLog) Debugf(format string, v ...interface{}) {
	l.messagef(LevelDebug, format, v...)
}

// Info logs some noticeable information
func (l *TestLog) Info(v ...interface{}) {
	l.message(LevelInfo, v...)
}

// Infof logs some noticeable information
func (l *TestLog) Infof(format string, v ...interface{}) {
	l.messagef(LevelInfo, format, v...)
}

// Warning send some important message to the console
func (l *TestLog) Warning(v ...interface{}) {
	l.message(LevelWarning, v...)
}

// Warningf send some important message to the console
func (l *TestLog) Warningf(format string, v ...interface{}) {
	l.messagef(LevelWarning, format, v...)
}

// Error sends error information to the console
func (l *TestLog) Error(v ...interface{}) {
	l.message(LevelError, v...)
}

// Errorf sends error information to the console
func (l *TestLog) Errorf(format string, v ...interface{}) {
	l.messagef(LevelError, format, v...)
}

func (l *TestLog) message(level LogLevel, v ...interface{}) {
	v = append([]interface{}{level.String()}, v...)
	l.t.Log(v...)
}

func (l *TestLog) messagef(level LogLevel, format string, v ...interface{}) {
	l.t.Logf(level.String()+" "+format, v...)
}

// Verify interface
var (
	_ Logger = &TestLog{}
)
