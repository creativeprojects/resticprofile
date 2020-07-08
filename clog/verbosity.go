package clog

// VerbosityMiddleware is a log middleware that is only passing log entries of level >= minimum level
type VerbosityMiddleware struct {
	destLog  Logger
	minLevel LogLevel
}

// NewVerbosityMiddleWare creates a new VerbosityMiddleware
// passing log entries to destination if level >= minimum level
func NewVerbosityMiddleWare(minLevel LogLevel, destination Logger) *VerbosityMiddleware {
	return &VerbosityMiddleware{
		minLevel: minLevel,
		destLog:  destination,
	}
}

// SetVerbosity changes the minimum level the log entries are going to be sent to the destination logger
func (l *VerbosityMiddleware) SetVerbosity(minLevel LogLevel) {
	l.minLevel = minLevel
}

// Log sends a log entry with the specified level
func (l *VerbosityMiddleware) Log(level LogLevel, v ...interface{}) {
	if level < l.minLevel {
		return
	}
	l.destLog.Log(level, v...)
}

// Logf sends a log entry with the specified level
func (l *VerbosityMiddleware) Logf(level LogLevel, format string, v ...interface{}) {
	if level < l.minLevel {
		return
	}
	l.destLog.Logf(level, format, v...)
}

// Debug sends debugging information
func (l *VerbosityMiddleware) Debug(v ...interface{}) {
	if DebugLevel < l.minLevel {
		return
	}
	l.destLog.Debug(v...)
}

// Debugf sends debugging information
func (l *VerbosityMiddleware) Debugf(format string, v ...interface{}) {
	if DebugLevel < l.minLevel {
		return
	}
	l.destLog.Debugf(format, v...)
}

// Info logs some noticeable information
func (l *VerbosityMiddleware) Info(v ...interface{}) {
	if InfoLevel < l.minLevel {
		return
	}
	l.destLog.Info(v...)
}

// Infof logs some noticeable information
func (l *VerbosityMiddleware) Infof(format string, v ...interface{}) {
	if InfoLevel < l.minLevel {
		return
	}
	l.destLog.Infof(format, v...)
}

// Warning send some important message to the console
func (l *VerbosityMiddleware) Warning(v ...interface{}) {
	if WarningLevel < l.minLevel {
		return
	}
	l.destLog.Warning(v...)
}

// Warningf send some important message to the console
func (l *VerbosityMiddleware) Warningf(format string, v ...interface{}) {
	if WarningLevel < l.minLevel {
		return
	}
	l.destLog.Warningf(format, v...)
}

// Error sends error information to the console
func (l *VerbosityMiddleware) Error(v ...interface{}) {
	// error level is always going through
	l.destLog.Error(v...)
}

// Errorf sends error information to the console
func (l *VerbosityMiddleware) Errorf(format string, v ...interface{}) {
	// error level is always going through
	l.destLog.Errorf(format, v...)
}

// Verify interface
var (
	_ Logger = &VerbosityMiddleware{}
)
