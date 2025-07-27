package ansible

import (
	cli "github.com/urfave/cli/v2"

	"github.com/bookandmusic/dev-tools/internal/plugin/base"
)

type AnsiblePlugin struct {
	meta     base.PluginMeta
	basePath string
}

func NewAnsiblePlugin(meta base.PluginMeta, basePath string) *AnsiblePlugin {
	return &AnsiblePlugin{meta: meta, basePath: basePath}
}

func (a *AnsiblePlugin) Name() string { return a.meta.Name }

func (a *AnsiblePlugin) Description() string { return a.meta.Description }

func (a *AnsiblePlugin) Command() *cli.Command {
	return base.CreateCommandTree(a.basePath, a.meta, &AnsibleExecutor{})
}
