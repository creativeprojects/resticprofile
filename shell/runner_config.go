package shell

import "io"

type RunnerConfig struct {
	Shell  Type
	Env    []string
	Dir    string
	Stdin  io.ReadCloser
	Stdout io.Writer
	Stderr io.Writer
}
