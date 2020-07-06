package clog

import (
	"log"

	"github.com/fatih/color"
)

// ConsoleLog logs messages to the console (in colour)
type ConsoleLog struct {
	quiet     bool
	verbose   bool
	colorMaps map[string][4]*color.Color
	levelMap  [4]*color.Color
}

// NewConsoleLog creates a new console logger
func NewConsoleLog() *ConsoleLog {
	console := &ConsoleLog{}
	console.init()
	return console
}

func (l *ConsoleLog) init() {
	l.colorMaps = map[string][4]*color.Color{
		"none": {
			DebugLevel:   nil,
			InfoLevel:    nil,
			WarningLevel: color.New(color.Bold),
			ErrorLevel:   color.New(color.Bold),
		},
		"light": {
			DebugLevel:   color.New(color.FgGreen),
			InfoLevel:    color.New(color.FgCyan),
			WarningLevel: color.New(color.FgMagenta, color.Bold),
			ErrorLevel:   color.New(color.FgRed, color.Bold),
		},
		"dark": {
			DebugLevel:   color.New(color.FgHiGreen),
			InfoLevel:    color.New(color.FgHiCyan),
			WarningLevel: color.New(color.FgHiMagenta, color.Bold),
			ErrorLevel:   color.New(color.FgHiRed, color.Bold),
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

// Quiet will only display warnings and errors
func (l *ConsoleLog) Quiet() {
	l.quiet = true
	l.verbose = false
}

// Verbose will display debugging information
func (l *ConsoleLog) Verbose() {
	l.verbose = true
	l.quiet = false
}

// Debug sends debugging information
func (l *ConsoleLog) Debug(v ...interface{}) {
	if !l.verbose {
		return
	}
	l.message(l.levelMap[DebugLevel], v...)
}

// Debugf sends debugging information
func (l *ConsoleLog) Debugf(format string, v ...interface{}) {
	if !l.verbose {
		return
	}
	l.messagef(l.levelMap[DebugLevel], format, v...)
}

// Info logs some noticeable information
func (l *ConsoleLog) Info(v ...interface{}) {
	if l.quiet {
		return
	}
	l.message(l.levelMap[InfoLevel], v...)
}

// Infof logs some noticeable information
func (l *ConsoleLog) Infof(format string, v ...interface{}) {
	if l.quiet {
		return
	}
	l.messagef(l.levelMap[InfoLevel], format, v...)
}

// Warning send some important message to the console
func (l *ConsoleLog) Warning(v ...interface{}) {
	l.message(l.levelMap[WarningLevel], v...)
}

// Warningf send some important message to the console
func (l *ConsoleLog) Warningf(format string, v ...interface{}) {
	l.messagef(l.levelMap[WarningLevel], format, v...)
}

// Error sends error information to the console
func (l *ConsoleLog) Error(v ...interface{}) {
	l.message(l.levelMap[ErrorLevel], v...)
}

// Errorf sends error information to the console
func (l *ConsoleLog) Errorf(format string, v ...interface{}) {
	l.messagef(l.levelMap[ErrorLevel], format, v...)
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
	_ Log = &ConsoleLog{}
)
