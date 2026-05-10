package main

import (
	"os"

	"github.com/sammyne/reddit-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
