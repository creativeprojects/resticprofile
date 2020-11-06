package main

import (
	_ "unsafe"
)

//go:linkname goarm runtime.goarm
var goarm uint8
