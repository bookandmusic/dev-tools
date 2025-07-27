package cmd

import (
	cli "github.com/urfave/cli/v2"

	"github.com/bookandmusic/dev-tools/internal/plugin"
)

// Version and BuildTime will be set during build
var (
	Version   string
	BuildTime string
)

// NewApp creates a new CLI app with all commands
func NewApp() *cli.App {
	app := &cli.App{
		Name:  "dev-tools",
		Usage: "A professional Go application",
	}
	plugin.LoadPluginsToRefistry()
	plugins := plugin.GetRegisteredPlugins()
	app.Commands = append(app.Commands, plugins...)
	return app
}
