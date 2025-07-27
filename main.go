package main

import (
	"os"

	"github.com/bookandmusic/dev-tools/cmd"
	_ "github.com/bookandmusic/dev-tools/internal/manager/soft/docker"
	_ "github.com/bookandmusic/dev-tools/internal/manager/soft/ohmyzsh"
	_ "github.com/bookandmusic/dev-tools/internal/manager/soft/self"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
