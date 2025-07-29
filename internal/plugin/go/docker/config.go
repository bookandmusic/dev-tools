package docker

import (
	"os"

	cli "github.com/urfave/cli/v2"

	"github.com/bookandmusic/dev-tools/internal/utils"
)

const (
	DefaultDockerVersion  = ""
	DefaultComposeVersion = ""
	DefaultBuildxVersion  = ""
	DefaultBinaryDomain   = "mirrors.tuna.tsinghua.edu.cn"
	DefaultProxy          = ""
	DefaultBinDir         = "/usr/local/bin"
	DefaultPluginDir      = "/usr/libexec/docker/cli-plugins"
	DefaultServicePath    = "/usr/lib/systemd/system"
)

func NewConfigFromContext(c *cli.Context, arch utils.ArchType) Config {
	return Config{
		Arch:           arch,
		DockerVersion:  valueOrDefault(c.String("docker-version"), DefaultDockerVersion),
		ComposeVersion: valueOrDefault(c.String("compose-version"), DefaultComposeVersion),
		BuildxVersion:  valueOrDefault(c.String("buildx-version"), DefaultBuildxVersion),
		Proxy:          valueOrDefault(c.String("proxy"), DefaultProxy),
		Domain:         valueOrDefault(c.String("domain"), DefaultBinaryDomain),
		BinDir:         os.ExpandEnv(valueOrDefault(c.String("bin-dir"), DefaultBinDir)),
		PluginDir:      os.ExpandEnv(valueOrDefault(c.String("plugin-dir"), DefaultPluginDir)),
		ServicePath:    os.ExpandEnv(valueOrDefault(c.String("service-path"), DefaultServicePath)),
	}
}

func valueOrDefault(val, def string) string {
	if val == "" {
		val = def
	}
	return val
}
