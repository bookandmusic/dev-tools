package loader

import (
	"os"
	"path/filepath"

	"github.com/bookandmusic/dev-tools/cmd/plugin"
	"github.com/bookandmusic/dev-tools/internal/manager/script"
	"github.com/bookandmusic/dev-tools/internal/ui"
)

func LoadPluginsFromLoader(rootDir string, ui ui.UI) {
	pluginDir := filepath.Join(rootDir, "plugins")

	info, err := os.Stat(pluginDir)
	if err != nil || !info.IsDir() {
		return
	}

	_ = filepath.Walk(pluginDir, func(path string, info os.FileInfo, err error) error {
		meta, metaErr := LoadPluginMeta(path, info)
		if metaErr != nil {
			return nil
		}
		switch meta.Type {
		case "shell":
			plugin.Register(CreateCommandTree(path, meta, script.NewShellExecutor(ui)))
		case "ansible":
			plugin.Register(CreateCommandTree(path, meta, script.NewAnsibleExecutor(ui)))
		}
		return nil
	})
}
