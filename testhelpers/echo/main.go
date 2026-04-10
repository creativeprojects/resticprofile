package main

import (
	"fmt"
	"os"
)

// displays all the parameters received on the command line
func main() {
	fmt.Printf("%q\n", os.Args[1:])
}
