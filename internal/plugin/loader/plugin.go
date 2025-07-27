package loader

import (
	"github.com/spf13/cobra"
)

type PluginExecutor interface {
	ScriptPath(basePath string, names ...string) string
	Exec(scriptPath string, cmd *cobra.Command, args []string) error
	NotFoundError(path string) error
}

type GenericPlugin struct {
	meta     PluginMeta
	basePath string
	executor PluginExecutor
}

func NewGenericPlugin(meta PluginMeta, basePath string, executor PluginExecutor) *GenericPlugin {
	return &GenericPlugin{
		meta:     meta,
		basePath: basePath,
		executor: executor,
	}
}

func (p *GenericPlugin) Name() string { return p.meta.Name }

func (p *GenericPlugin) Description() string { return p.meta.Description }

func (p *GenericPlugin) Command() *cobra.Command {
	return CreateCommandTree(p.basePath, p.meta, p.executor)
}
