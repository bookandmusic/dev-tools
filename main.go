package main

import (
	"os"

	"github.com/bookandmusic/dev-tools/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
