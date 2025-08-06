package main

import (
	"log"
	"os"

	"github.com/bookandmusic/dev-tools/cmd"
	"github.com/bookandmusic/dev-tools/internal/config"
	"github.com/bookandmusic/dev-tools/internal/ui"
)

func main() {
	ui.NewConsoleUI()
	config.LoadConfig()
	app := cmd.NewApp()

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
