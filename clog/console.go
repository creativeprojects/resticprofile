package clog

import (
	"log"

	"github.com/fatih/color"
)

const (
	numLevels = 4 // Debug, Info, Warn, Error
)

// ConsoleLog logs messages to the console (in colour)
type ConsoleLog struct {
	colorMaps map[string][numLevels]*color.Color
	levelMap  [numLevels]*color.Color
}

// NewConsoleLog creates a new console logger
func NewConsoleLog() *ConsoleLog {
	console := &ConsoleLog{}
	console.init()
	return console
}

func (l *ConsoleLog) init() {
	l.colorMaps = map[string][numLevels]*color.Color{
		"none": {
			LevelDebug:   nil,
			LevelInfo:    nil,
			LevelWarning: color.New(color.Bold),
			LevelError:   color.New(color.Bold),
		},
		"light": {
			LevelDebug:   color.New(color.FgGreen),
			LevelInfo:    color.New(color.FgCyan),
			LevelWarning: color.New(color.FgMagenta, color.Bold),
			LevelError:   color.New(color.FgRed, color.Bold),
		},
		"dark": {
			LevelDebug:   color.New(color.FgHiGreen),
			LevelInfo:    color.New(color.FgHiCyan),
			LevelWarning: color.New(color.FgHiMagenta, color.Bold),
			LevelError:   color.New(color.FgHiRed, color.Bold),
		},
	}
	l.levelMap = l.colorMaps["light"]
}

// SetTheme sets the dark or light theme
func (l *ConsoleLog) SetTheme(theme string) {
	var ok bool
	l.levelMap, ok = l.colorMaps[theme]
	if !ok {
		l.levelMap = l.colorMaps["none"]
	}
}

// Colorize activate of deactivate colouring
func (l *ConsoleLog) Colorize(colorize bool) {
	color.NoColor = !colorize
}

// Log sends a log entry with the specified level
func (l *ConsoleLog) Log(level LogLevel, v ...interface{}) {
	l.message(l.levelMap[level], v...)
}

// Logf sends a log entry with the specified level
func (l *ConsoleLog) Logf(level LogLevel, format string, v ...interface{}) {
	l.messagef(l.levelMap[level], format, v...)
}

// Debug sends debugging information
func (l *ConsoleLog) Debug(v ...interface{}) {
	l.message(l.levelMap[LevelDebug], v...)
}

// Debugf sends debugging information
func (l *ConsoleLog) Debugf(format string, v ...interface{}) {
	l.messagef(l.levelMap[LevelDebug], format, v...)
}

// Info logs some noticeable information
func (l *ConsoleLog) Info(v ...interface{}) {
	l.message(l.levelMap[LevelInfo], v...)
}

// Infof logs some noticeable information
func (l *ConsoleLog) Infof(format string, v ...interface{}) {
	l.messagef(l.levelMap[LevelInfo], format, v...)
}

// Warning send some important message to the console
func (l *ConsoleLog) Warning(v ...interface{}) {
	l.message(l.levelMap[LevelWarning], v...)
}

// Warningf send some important message to the console
func (l *ConsoleLog) Warningf(format string, v ...interface{}) {
	l.messagef(l.levelMap[LevelWarning], format, v...)
}

// Error sends error information to the console
func (l *ConsoleLog) Error(v ...interface{}) {
	l.message(l.levelMap[LevelError], v...)
}

// Errorf sends error information to the console
func (l *ConsoleLog) Errorf(format string, v ...interface{}) {
	l.messagef(l.levelMap[LevelError], format, v...)
}

func (l *ConsoleLog) message(c *color.Color, v ...interface{}) {
	l.setColor(c)
	log.Println(v...)
	l.unsetColor()
}

func (l *ConsoleLog) messagef(c *color.Color, format string, v ...interface{}) {
	l.setColor(c)
	log.Printf(format+"\n", v...)
	l.unsetColor()
}

func (l *ConsoleLog) setColor(c *color.Color) {
	if c != nil {
		c.Set()
	}
}

func (l *ConsoleLog) unsetColor() {
	color.Unset()
}

// Verify interface
var (
	_ Logger = &ConsoleLog{}
)
