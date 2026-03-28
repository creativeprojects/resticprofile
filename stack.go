package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

// getStack returns a simplified stack trace
func getStack(skip int) string {
	stack := new(strings.Builder)
	pc := make([]uintptr, 20)        // stack of 20 traces max
	n := runtime.Callers(skip+1, pc) // skip call to runtime.Callers
	if n == 0 {
		return ""
	}

	pc = pc[:n] // pass only valid pcs to runtime.CallersFrames
	frames := runtime.CallersFrames(pc)

	// Loop to get frames.
	// A fixed number of PCs can expand to an indefinite number of Frames.
	for {
		frame, more := frames.Next()

		if frame.Function == "runtime.main" {
			// last 2 traces are inside the go bootstrap
			break
		}

		fmt.Fprintf(stack, "%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)

		if !more {
			break
		}
	}
	if stack.String() == "" {
		// the stack trace is suspiciously empty, let's try another way instead
		return string(debug.Stack())
	}
	return stack.String()
}
