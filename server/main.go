package main

import (
	"os"

	"github.com/chaitanyamaili/go-rpc-cloud-build/server/cmd"
)

func main() {
	// Try to execute, or close if an error was found
	err := cmd.Execute()
	if err != nil {
		os.Exit(-1)
	}
}
