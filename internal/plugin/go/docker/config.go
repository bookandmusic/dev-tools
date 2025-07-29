package docker

import (
	cli "github.com/urfave/cli/v2"

	"github.com/bookandmusic/dev-tools/internal/utils"
)

const (
	DefaultDockerVersion  = "28.3.2"
	DefaultComposeVersion = "2.39.1"
	DefaultBuildxVersion  = "0.26.1"
	DefaultProxy          = "https://gitproxy.click"
	DefaultBinDir         = "/usr/local/bin"
	DefaultPluginDir      = "/usr/local/libexec/docker/cli-plugins"
	DefaultServicePath    = "/etc/systemd/system"
)

func NewConfigFromContext(c *cli.Context, arch utils.ArchType) Config {
	return Config{
		Arch:           arch,
		DockerVersion:  valueOrDefault(c.String("docker-version"), DefaultDockerVersion),
		ComposeVersion: valueOrDefault(c.String("compose-version"), DefaultComposeVersion),
		BuildxVersion:  valueOrDefault(c.String("buildx-version"), DefaultBuildxVersion),
		Proxy:          valueOrDefault(c.String("proxy"), DefaultProxy),
		BinDir:         valueOrDefault(c.String("bin-dir"), DefaultBinDir),
		PluginDir:      valueOrDefault(c.String("plugin-dir"), DefaultPluginDir),
		ServicePath:    valueOrDefault(c.String("service-path"), DefaultServicePath),
	}
}

func valueOrDefault(val, def string) string {
	if val != "" {
		return val
	}
	return def
}
