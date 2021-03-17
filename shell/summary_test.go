package shell

import "runtime"

var (
	eol = "\n"
)

func init() {
	if runtime.GOOS == "windows" {
		eol = "\r\n"
	}
}
