package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Environment variables:")
	for _, value := range os.Environ() {
		fmt.Println(value)
	}

	cwd, _ := os.Getwd()
	fmt.Println("\nCurrent directory", cwd)
}
