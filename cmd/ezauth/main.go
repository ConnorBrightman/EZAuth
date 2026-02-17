package main

import (
	"fmt"
	"os"
)

var Version = "dev"

func main() {
	Execute()
}

func Execute() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	switch os.Args[1] {

	case "init":
		runInit()

	case "start":
		runStart()

	case "version", "--version", "-v":
		fmt.Println("ezauth version:", Version)

	default:
		fmt.Printf("Unknown command: %s\n\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}
