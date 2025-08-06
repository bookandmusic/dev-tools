package goplugin

import (
	cli "github.com/urfave/cli/v2"

	"github.com/bookandmusic/dev-tools/internal/plugin/go/docker"
	"github.com/bookandmusic/dev-tools/internal/ui"
	"github.com/bookandmusic/dev-tools/internal/utils"
)

type DockerPlugin struct{}

func (g *DockerPlugin) Name() string {
	return "docker"
}

func (g *DockerPlugin) Description() string {
	return "A Go plugin for managing Docker"
}

var commonFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  "bin-dir",
		Usage: "Binary directory",
		Value: docker.DefaultBinDir,
	},
	&cli.StringFlag{
		Name:  "plugin-dir",
		Usage: "Docker CLI plugin directory",
		Value: docker.DefaultPluginDir,
	},
	&cli.StringFlag{
		Name:  "service-path",
		Usage: "systemd service directory",
		Value: docker.DefaultServicePath,
	},
}

var installFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  "proxy",
		Usage: "Optional proxy to GitHub",
		Value: docker.DefaultProxy,
	},
	&cli.StringFlag{
		Name:  "domain",
		Usage: "Optional domain to Docker static binary: mirrors.tuna.tsinghua.edu.cn, mirrors.ustc.edu.cn, mirrors.aliyun.com",
		Value: docker.DefaultBinaryDomain,
	},
	&cli.StringFlag{
		Name:  "docker-version",
		Usage: "Docker version",
		Value: docker.DefaultDockerVersion,
	},
	&cli.StringFlag{
		Name:  "compose-version",
		Usage: "Compose plugin version",
		Value: docker.DefaultComposeVersion,
	},
	&cli.StringFlag{
		Name:  "buildx-version",
		Usage: "Buildx plugin version",
		Value: docker.DefaultBuildxVersion,
	},
}

func (g *DockerPlugin) Command() *cli.Command {
	var manager docker.DockerManager
	return &cli.Command{
		Name:  g.Name(),
		Usage: g.Description(),
		Before: func(c *cli.Context) error {
			if manager != nil {
				return nil
			}
			arch := utils.DetectArch()
			if arch == utils.ArchUnknown {
				return cli.Exit("Unsupported architecture", 1)
			}
			config := docker.NewConfigFromContext(c, arch)
			ui := ui.ConsoleUI{}
			manager = docker.New(config, ui)
			return nil
		},
		Subcommands: []*cli.Command{
			{
				Name:    "install",
				Aliases: []string{"i"},
				Usage:   "Install Docker",
				Flags:   append(installFlags, commonFlags...),
				Action: func(c *cli.Context) error {
					return manager.Install()
				},
			},
			{
				Name:    "uninstall",
				Aliases: []string{"u"},
				Usage:   "Uninstall Docker and related components",
				Flags:   commonFlags,
				Action: func(c *cli.Context) error {
					return manager.Uninstall()
				},
			},
		},
	}
}
