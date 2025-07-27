// plugin/registry.go
package plugin

import (
	"os"
	"path/filepath"

	cli "github.com/urfave/cli/v2"

	"github.com/bookandmusic/dev-tools/internal/plugin/ansible"
	"github.com/bookandmusic/dev-tools/internal/plugin/base"
	goPlugin "github.com/bookandmusic/dev-tools/internal/plugin/go"
	"github.com/bookandmusic/dev-tools/internal/plugin/shell"
)

// PluginApp 插件统一接口
type PluginApp interface {
	Name() string
	Description() string
	Command() *cli.Command
}

var (
	pluginApps []PluginApp
)

// RegisterPluginApp 注册插件实现
func RegisterPluginApp(app PluginApp) {
	pluginApps = append(pluginApps, app)
}

// GetRegisteredPlugins 返回所有注册的 CLI 命令
func GetRegisteredPlugins() []*cli.Command {
	var cmds []*cli.Command
	for _, app := range pluginApps {
		cmds = append(cmds, app.Command())
	}
	return cmds
}

func LoadPluginsToRefistry() {
	// 外部插件
	pluginDir := "plugins"
	_ = filepath.Walk(pluginDir, func(path string, info os.FileInfo, err error) error {
		meta, meatErr := base.LoadPluginMeta(path, info)
		if meatErr != nil {
			return nil
		}
		switch meta.Type {
		case "shell":
			RegisterPluginApp(shell.NewShellPlugin(
				*meta,
				path,
			))
		case "ansible":
			RegisterPluginApp(ansible.NewAnsiblePlugin(
				*meta,
				path,
			))
		}
		return nil
	})
	// go插件
	RegisterPluginApp(&goPlugin.GoPlugin{})
}
