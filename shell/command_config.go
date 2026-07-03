package shell

import "io"

type CommandConfig struct {
	Command    string
	Args       []string
	PublicArgs []string
	Stdin      io.ReadCloser
	Stdout     io.Writer
	Stderr     io.Writer
	SetPID     SetPID
}
