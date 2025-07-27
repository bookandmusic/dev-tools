package shell

import (
	cli "github.com/urfave/cli/v2"

	"github.com/bookandmusic/dev-tools/internal/plugin/base"
)

type ShellPlugin struct {
	meta     base.PluginMeta
	basePath string
}

func NewShellPlugin(meta base.PluginMeta, basePath string) *ShellPlugin {
	return &ShellPlugin{meta: meta, basePath: basePath}
}

func (s *ShellPlugin) Name() string { return s.meta.Name }

func (s *ShellPlugin) Description() string { return s.meta.Description }

func (s *ShellPlugin) Command() *cli.Command {
	return base.CreateCommandTree(s.basePath, s.meta, &ShellExecutor{})
}
