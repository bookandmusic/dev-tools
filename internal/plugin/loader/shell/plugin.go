package shell

import (
	"github.com/spf13/cobra"

	"github.com/bookandmusic/dev-tools/internal/plugin/loader"
)

type ShellPlugin struct {
	meta     loader.PluginMeta
	basePath string
}

func NewShellPlugin(meta loader.PluginMeta, basePath string) *ShellPlugin {
	return &ShellPlugin{meta: meta, basePath: basePath}
}

func (s *ShellPlugin) Name() string { return s.meta.Name }

func (s *ShellPlugin) Description() string { return s.meta.Description }

func (s *ShellPlugin) Command() *cobra.Command {
	return loader.CreateCommandTree(s.basePath, s.meta, &ShellExecutor{})
}
