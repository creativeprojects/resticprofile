package clog

// LevelFilter is a log middleware that is only passing log entries of level >= minimum level
type LevelFilter struct {
	destLog  Logger
	minLevel LogLevel
}

// NewLevelFilter creates a new VerbosityMiddleware
// passing log entries to destination if level >= minimum level
func NewLevelFilter(minLevel LogLevel, destination Logger) *LevelFilter {
	return &LevelFilter{
		minLevel: minLevel,
		destLog:  destination,
	}
}

// SetVerbosity changes the minimum level the log entries are going to be sent to the destination logger
func (l *LevelFilter) SetVerbosity(minLevel LogLevel) {
	l.minLevel = minLevel
}

// Log sends a log entry with the specified level
func (l *LevelFilter) Log(level LogLevel, v ...interface{}) {
	if level < l.minLevel {
		return
	}
	l.destLog.Log(level, v...)
}

// Logf sends a log entry with the specified level
func (l *LevelFilter) Logf(level LogLevel, format string, v ...interface{}) {
	if level < l.minLevel {
		return
	}
	l.destLog.Logf(level, format, v...)
}

// Debug sends debugging information
func (l *LevelFilter) Debug(v ...interface{}) {
	if LevelDebug < l.minLevel {
		return
	}
	l.destLog.Debug(v...)
}

// Debugf sends debugging information
func (l *LevelFilter) Debugf(format string, v ...interface{}) {
	if LevelDebug < l.minLevel {
		return
	}
	l.destLog.Debugf(format, v...)
}

// Info logs some noticeable information
func (l *LevelFilter) Info(v ...interface{}) {
	if LevelInfo < l.minLevel {
		return
	}
	l.destLog.Info(v...)
}

// Infof logs some noticeable information
func (l *LevelFilter) Infof(format string, v ...interface{}) {
	if LevelInfo < l.minLevel {
		return
	}
	l.destLog.Infof(format, v...)
}

// Warning send some important message to the console
func (l *LevelFilter) Warning(v ...interface{}) {
	if LevelWarning < l.minLevel {
		return
	}
	l.destLog.Warning(v...)
}

// Warningf send some important message to the console
func (l *LevelFilter) Warningf(format string, v ...interface{}) {
	if LevelWarning < l.minLevel {
		return
	}
	l.destLog.Warningf(format, v...)
}

// Error sends error information to the console
func (l *LevelFilter) Error(v ...interface{}) {
	// error level is always going through
	l.destLog.Error(v...)
}

// Errorf sends error information to the console
func (l *LevelFilter) Errorf(format string, v ...interface{}) {
	// error level is always going through
	l.destLog.Errorf(format, v...)
}

// Verify interface
var (
	_ Logger = &LevelFilter{}
)
