package goplugin

import (
	"fmt"

	cli "github.com/urfave/cli/v2"
)

type GoPlugin struct{}

// Name 返回插件名称
func (g *GoPlugin) Name() string {
	return "example-go"
}

// Description 返回插件描述
func (g *GoPlugin) Description() string {
	return "An example shell plugin for testing"
}

// Command 返回子命令
func (g *GoPlugin) Command() *cli.Command {
	return &cli.Command{
		Name:  g.Name(),
		Usage: g.Description(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "message",
				Usage: "Optional message to print",
			},
		},
		Action: func(c *cli.Context) error {
			msg := c.String("message")
			if msg == "" {
				msg = "Hello from internal Go plugin!"
			}
			fmt.Println(msg)
			return nil
		},
	}
}
