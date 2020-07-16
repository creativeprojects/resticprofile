package clog

// LogLevel represents the importance of a log entry
type LogLevel int

// LogLevel
const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarning
	LevelError
)

// String representation of a level (5 characters)
func (level LogLevel) String() string {
	switch level {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO "
	case LevelWarning:
		return "WARN "
	case LevelError:
		return "ERROR"
	default:
		return "     "
	}
}
