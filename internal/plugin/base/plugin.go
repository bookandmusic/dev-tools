// plugin/base_plugin.go
package base

import cli "github.com/urfave/cli/v2"

type PluginExecutor interface {
	ScriptPath(basePath string, names ...string) string
	Exec(scriptPath string, c *cli.Context) error
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

func (p *GenericPlugin) Command() *cli.Command {
	return CreateCommandTree(p.basePath, p.meta, p.executor)
}
