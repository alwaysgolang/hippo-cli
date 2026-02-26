package main

import (
	"fmt"
	"hippo-cli/cmd/internal/build"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: hippo build [--verbose]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "build":
		verbose := false
		for _, a := range os.Args[2:] {
			if a == "--verbose" || a == "-v" {
				verbose = true
			}
		}

		if err := build.Run(build.Options{Verbose: verbose}); err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
	default:
		fmt.Println("unknown command")
		os.Exit(1)
	}
}
