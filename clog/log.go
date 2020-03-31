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
	colorMaps map[string]([4]*color.Color)
	levelMap  [4]*color.Color
	bold      *color.Color
)

func init() {
	bold = color.New(color.Bold)
	colorMaps = map[string]([4]*color.Color){
		"none": [4]*color.Color{
			DebugLevel:   nil,
			InfoLevel:    nil,
			WarningLevel: nil,
			ErrorLevel:   nil,
		},
		"light": [4]*color.Color{
			DebugLevel:   color.New(color.FgGreen),
			InfoLevel:    color.New(color.FgCyan),
			WarningLevel: color.New(color.FgMagenta),
			ErrorLevel:   color.New(color.FgRed),
		},
		"dark": [4]*color.Color{
			DebugLevel:   color.New(color.FgHiGreen),
			InfoLevel:    color.New(color.FgHiCyan),
			WarningLevel: color.New(color.FgHiMagenta),
			ErrorLevel:   color.New(color.FgHiRed),
		},
	}
	levelMap = colorMaps["light"]
}

func SetTheme(theme string) {
	var ok bool
	levelMap, ok = colorMaps[theme]
	if !ok {
		levelMap = colorMaps["none"]
	}
}

func Colorize(colorize bool) {
	color.NoColor = !colorize
}

func SetLevel(quietFlag, verboseFlag bool) {
	quiet = quietFlag
	verbose = verboseFlag
}

func Debug(v ...interface{}) {
	if !verbose {
		return
	}
	message(levelMap[DebugLevel], v...)
}

func Debugf(format string, v ...interface{}) {
	if !verbose {
		return
	}
	messagef(levelMap[DebugLevel], format, v...)
}

func Info(v ...interface{}) {
	if quiet {
		return
	}
	message(levelMap[InfoLevel], v...)
}

func Infof(format string, v ...interface{}) {
	if quiet {
		return
	}
	messagef(levelMap[InfoLevel], format, v...)
}

func Warning(v ...interface{}) {
	setBold()
	message(levelMap[WarningLevel], v...)
}

func Warningf(format string, v ...interface{}) {
	setBold()
	messagef(levelMap[WarningLevel], format, v...)
}

func Error(v ...interface{}) {
	setBold()
	message(levelMap[ErrorLevel], v...)
}

func Errorf(format string, v ...interface{}) {
	setBold()
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

func setBold() {
	bold.Set()
}
func unsetColor() {
	color.Unset()
}
