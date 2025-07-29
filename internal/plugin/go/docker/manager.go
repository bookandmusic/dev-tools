package docker

import (
	"github.com/bookandmusic/dev-tools/internal/ui"
	"github.com/bookandmusic/dev-tools/internal/utils"
)

type DockerManager interface {
	Install() error
	Uninstall() error
}

type Config struct {
	DockerVersion  string
	BuildxVersion  string
	ComposeVersion string
	Proxy          string
	Domain         string
	Arch           utils.ArchType
	BinDir         string
	PluginDir      string
	ServicePath    string
}

type installer struct {
	cfg Config
	ui  ui.UI
}

func New(cfg Config, ui ui.UI) DockerManager {
	return &installer{cfg: cfg, ui: ui}
}
