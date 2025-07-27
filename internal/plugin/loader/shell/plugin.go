package shell

import (
	"github.com/spf13/cobra"

	"github.com/bookandmusic/dev-tools/internal/plugin/loader"
	"github.com/bookandmusic/dev-tools/internal/ui"
)

type ShellPlugin struct {
	meta     loader.PluginMeta
	basePath string
	ui       ui.UI
}

func NewShellPlugin(meta loader.PluginMeta, basePath string, ui ui.UI) *ShellPlugin {
	return &ShellPlugin{meta: meta, basePath: basePath, ui: ui}
}

func (s *ShellPlugin) Name() string { return s.meta.Name }

func (s *ShellPlugin) Description() string { return s.meta.Description }

func (s *ShellPlugin) Command() *cobra.Command {
	return loader.CreateCommandTree(s.basePath, s.meta, NewShellExecutor(s.ui))
}
