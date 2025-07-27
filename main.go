package main

import (
	"log"
	"os"

	"github.com/bookandmusic/dev-tools/cmd"
)

func main() {
	app := cmd.NewApp()

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
