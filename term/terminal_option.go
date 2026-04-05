package term

import (
	"io"
	"os"

	"github.com/creativeprojects/resticprofile/util"
	"github.com/creativeprojects/resticprofile/util/maybe"
)

type terminalOption func(t *Terminal)

func WithStdin(stdin io.Reader) terminalOption {
	if stdin == nil {
		stdin = nilReader{}
	}
	return func(t *Terminal) {
		t.stdin = stdin
	}
}

func WithStdout(stdout io.Writer) terminalOption {
	if stdout == nil {
		stdout = io.Discard
	}
	if stdout != os.Stdout && stdout != os.Stderr {
		stdout = util.NewSyncWriter(stdout)
	}
	return func(t *Terminal) {
		t.stdout = stdout
	}
}

func WithStderr(stderr io.Writer) terminalOption {
	if stderr == nil {
		stderr = io.Discard
	}
	if stderr != os.Stdout && stderr != os.Stderr {
		stderr = util.NewSyncWriter(stderr)
	}
	return func(t *Terminal) {
		t.stderr = stderr
	}
}

func WithColors(enable bool) terminalOption {
	return func(t *Terminal) {
		t.enableColors = maybe.SetBool(enable)
	}
}

func WithStdoutRecorder(recorder *Recorder) terminalOption {
	return func(t *Terminal) {
		t.stdout = recorder.inputWriter
	}
}

func WithStderrRecorder(recorder *Recorder) terminalOption {
	return func(t *Terminal) {
		t.stderr = recorder.inputWriter
	}
}
