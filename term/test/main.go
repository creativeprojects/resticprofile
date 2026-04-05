package main

import (
	"github.com/creativeprojects/resticprofile/term"
	"github.com/creativeprojects/resticprofile/util/ansi"
)

func main() {
	terminal := term.NewTerminal()
	terminal.Println(ansi.Bold("colorable terminal"))
	terminal.Printf("stdout is terminal: %v, stderr is terminal: %v\n", terminal.StdoutIsTerminal(), terminal.StderrIsTerminal())

	terminal = term.NewTerminal(term.WithColors(false))
	terminal.Println(ansi.Bold("non colorable terminal"))
	terminal.Printf("stdout is terminal: %v, stderr is terminal: %v\n", terminal.StdoutIsTerminal(), terminal.StderrIsTerminal())
}
