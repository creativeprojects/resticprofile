package main

import (
	"flag"
	"io"
	"log"
	"os"
)

func main() {
	var list bool
	flag.BoolVar(&list, "l", false, "list")
	flag.Parse()
	if list {
		return
	}
	_, err := io.Copy(os.Stdout, os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
}
