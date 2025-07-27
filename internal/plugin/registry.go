// plugin/registry.go
package plugin

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/bookandmusic/dev-tools/internal/config"
	"github.com/bookandmusic/dev-tools/internal/plugin/adapter"
	"github.com/bookandmusic/dev-tools/internal/plugin/loader"
	"github.com/bookandmusic/dev-tools/internal/plugin/loader/ansible"
	"github.com/bookandmusic/dev-tools/internal/plugin/loader/shell"
	"github.com/bookandmusic/dev-tools/internal/ui"
)

// PluginApp 插件统一接口
type PluginApp interface {
	Name() string
	Description() string
	Command() *cobra.Command
}

var pluginApps []PluginApp

// RegisterPluginApp 注册插件实现
func RegisterPluginApp(app PluginApp) {
	pluginApps = append(pluginApps, app)
}

// GetRegisteredPlugins 返回所有注册的 CLI 命令
func GetRegisteredPlugins() []*cobra.Command {
	var cmds []*cobra.Command
	for _, app := range pluginApps {
		cmds = append(cmds, app.Command())
	}
	return cmds
}

func LoadPluginsToRefistry(rootDir string, ui ui.UI, cfg *config.GlobalConfig) {
	pluginDir := "plugins" // 默认的相对路径
	pluginDir = filepath.Join(rootDir, pluginDir)

	// 检查插件目录是否存在
	info, err := os.Stat(pluginDir)
	if err == nil && info.IsDir() {
		// 遍历插件目录
		_ = filepath.Walk(pluginDir, func(path string, info os.FileInfo, err error) error {
			meta, metaErr := loader.LoadPluginMeta(path, info)
			if metaErr != nil {
				return nil
			}
			switch meta.Type {
			case "shell":
				RegisterPluginApp(shell.NewShellPlugin(
					*meta,
					path,
					ui,
				))
			case "ansible":
				RegisterPluginApp(ansible.NewAnsiblePlugin(
					*meta,
					path,
					ui,
				))
			}
			return nil
		})
	}
	// 注册内置插件
	RegisterPluginApp(adapter.NewSelfPlugin(ui, cfg))
	RegisterPluginApp(adapter.NewOhMyzshPlugin(ui, cfg))
	RegisterPluginApp(adapter.NewDockerPlugin(ui, cfg))
}
