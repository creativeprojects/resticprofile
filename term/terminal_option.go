package term

import (
	"io"
	"os"

	"github.com/creativeprojects/resticprofile/util"
	"github.com/creativeprojects/resticprofile/util/maybe"
)

type TerminalOption func(t *Terminal)

func WithNoStdin() TerminalOption {
	return func(t *Terminal) {
		t.stdin = nilReader{}
	}
}

func WithNoStdout() TerminalOption {
	return func(t *Terminal) {
		t.stdout = io.Discard
	}
}

func WithNoStderr() TerminalOption {
	return func(t *Terminal) {
		t.stderr = io.Discard
	}
}

func WithStdin(stdin io.Reader) TerminalOption {
	if stdin == nil {
		return func(t *Terminal) {}
	}
	return func(t *Terminal) {
		t.stdin = stdin
	}
}

func WithStdout(stdout io.Writer) TerminalOption {
	if stdout == nil {
		return func(t *Terminal) {}
	}
	if stdout != os.Stdout && stdout != os.Stderr {
		stdout = util.NewSyncWriter(stdout)
	}
	return func(t *Terminal) {
		t.stdout = stdout
	}
}

func WithStderr(stderr io.Writer) TerminalOption {
	if stderr == nil {
		return func(t *Terminal) {}
	}
	if stderr != os.Stdout && stderr != os.Stderr {
		stderr = util.NewSyncWriter(stderr)
	}
	return func(t *Terminal) {
		t.stderr = stderr
	}
}

func WithColors(enable bool) TerminalOption {
	return func(t *Terminal) {
		t.enableColors = maybe.SetBool(enable)
	}
}

func WithStdoutRecorder(recorder *Recorder) TerminalOption {
	return func(t *Terminal) {
		t.stdout = recorder.inputWriter
	}
}

func WithStderrRecorder(recorder *Recorder) TerminalOption {
	return func(t *Terminal) {
		t.stderr = recorder.inputWriter
	}
}

// WithStdoutCopy creates a copy of everything sent to Stdout to `copy` writer.
// Colorisation is independent: the Stdout writer can display colors while the copy to a non terminal won't display colors.
func WithStdoutCopy(w io.Writer) TerminalOption {
	return func(t *Terminal) {
		t.copyStdout = w
	}
}

func WithStderrCopy(w io.Writer) TerminalOption {
	return func(t *Terminal) {
		t.copyStderr = w
	}
}
