package ansible

import (
	"github.com/spf13/cobra"

	"github.com/bookandmusic/dev-tools/internal/plugin/loader"
)

type AnsiblePlugin struct {
	meta     loader.PluginMeta
	basePath string
}

func NewAnsiblePlugin(meta loader.PluginMeta, basePath string) *AnsiblePlugin {
	return &AnsiblePlugin{meta: meta, basePath: basePath}
}

func (a *AnsiblePlugin) Name() string { return a.meta.Name }

func (a *AnsiblePlugin) Description() string { return a.meta.Description }

func (a *AnsiblePlugin) Command() *cobra.Command {
	return loader.CreateCommandTree(a.basePath, a.meta, &AnsibleExecutor{})
}
