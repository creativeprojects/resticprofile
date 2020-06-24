package clog

import (
	"log"

	"github.com/fatih/color"
)

// LogLevel
const (
	DebugLevel = iota
	InfoLevel
	WarningLevel
	ErrorLevel
)

var (
	quiet     bool
	verbose   bool
	colorMaps map[string][4]*color.Color
	levelMap  [4]*color.Color
)

func init() {
	colorMaps = map[string][4]*color.Color{
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
	levelMap = colorMaps["light"]
}

// SetTheme sets the dark or light theme
func SetTheme(theme string) {
	var ok bool
	levelMap, ok = colorMaps[theme]
	if !ok {
		levelMap = colorMaps["none"]
	}
}

// Colorize activate of deactivate colouring
func Colorize(colorize bool) {
	color.NoColor = !colorize
}

// Quiet will only display warnings and errors
func Quiet() {
	quiet = true
	verbose = false
}

// Verbose will display debugging information
func Verbose() {
	verbose = true
	quiet = false
}

// Debug sends debugging information
func Debug(v ...interface{}) {
	if !verbose {
		return
	}
	message(levelMap[DebugLevel], v...)
}

// Debugf sends debugging information
func Debugf(format string, v ...interface{}) {
	if !verbose {
		return
	}
	messagef(levelMap[DebugLevel], format, v...)
}

// Info logs some noticeable information
func Info(v ...interface{}) {
	if quiet {
		return
	}
	message(levelMap[InfoLevel], v...)
}

// Infof logs some noticeable information
func Infof(format string, v ...interface{}) {
	if quiet {
		return
	}
	messagef(levelMap[InfoLevel], format, v...)
}

// Warning send some important message to the console
func Warning(v ...interface{}) {
	message(levelMap[WarningLevel], v...)
}

// Warningf send some important message to the console
func Warningf(format string, v ...interface{}) {
	messagef(levelMap[WarningLevel], format, v...)
}

// Error sends error information to the console
func Error(v ...interface{}) {
	message(levelMap[ErrorLevel], v...)
}

// Errorf sends error information to the console
func Errorf(format string, v ...interface{}) {
	messagef(levelMap[ErrorLevel], format, v...)
}

func message(c *color.Color, v ...interface{}) {
	setColor(c)
	log.Println(v...)
	unsetColor()
}

func messagef(c *color.Color, format string, v ...interface{}) {
	setColor(c)
	log.Printf(format+"\n", v...)
	unsetColor()
}

func setColor(c *color.Color) {
	if c != nil {
		c.Set()
	}
}

func unsetColor() {
	color.Unset()
}
