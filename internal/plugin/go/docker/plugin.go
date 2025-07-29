package docker

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bookandmusic/dev-tools/internal/utils"
)

// installPlugins 安装 CLI 插件（Compose / Buildx）
func (i *installer) installPlugins() error {
	i.ui.Info("Installing Docker CLI plugins to %s...", i.cfg.PluginDir)

	// 1. 确保插件目录存在
	if err := utils.CreateDirIfNotExists(i.cfg.PluginDir); err != nil {
		i.ui.Error("Failed to create plugin directory %s: %v", i.cfg.PluginDir, err)
		return fmt.Errorf("failed to create directory: %w", err)
	}
	i.ui.Success("Plugin directory ready: %s", i.cfg.PluginDir)

	// 2. 定义要安装的插件列表
	plugins := []struct {
		Name string
		URL  string
	}{
		{
			Name: "docker-compose",
			URL: fmt.Sprintf("https://github.com/docker/compose/releases/download/%s/docker-compose-linux-%s",
				i.cfg.ComposeVersion, i.cfg.Arch),
		},
		{
			Name: "docker-buildx",
			URL: fmt.Sprintf("https://github.com/docker/buildx/releases/download/%s/buildx-%s.linux-%s",
				i.cfg.BuildxVersion, i.cfg.BuildxVersion, map[utils.ArchType]string{
					utils.ArchAARCH64: "arm64",
					utils.ArchX86_64:  "amd64",
				}[i.cfg.Arch]),
		},
	}

	// 3. 顺序安装插件
	for _, p := range plugins {
		if err := i.installPlugin(p.Name, p.URL); err != nil {
			return err
		}
	}

	i.ui.Success("All Docker CLI plugins installed successfully.")
	return nil
}

// installPlugin 下载并安装单个插件
func (i *installer) installPlugin(pluginName, url string) error {
	if i.cfg.Proxy != "" {
		url = i.cfg.Proxy + "/" + url
	}

	pluginPath := filepath.Join(i.cfg.PluginDir, pluginName)
	i.ui.Info("Installing plugin: %s", pluginName)

	// 1. 下载插件
	i.ui.Info("Downloading plugin from %s to %s...", url, pluginPath)
	if err := utils.DownloadFileWithProgress(url, pluginPath, i.ui); err != nil {
		i.ui.Error("Failed to download plugin %s: %v", pluginName, err)
		return err
	}
	i.ui.Success("Plugin %s downloaded successfully.", pluginName)

	// 2. 添加执行权限
	i.ui.Info("Adding execute permission to plugin %s...", pluginName)
	if err := utils.AddExecPermission(pluginPath); err != nil {
		i.ui.Error("Failed to add execute permission to plugin %s: %v", pluginName, err)
		return err
	}

	i.ui.Success("Plugin %s installed successfully.", pluginName)
	return nil
}

// ensureClientConfig 确保 ~/.docker/config.json 包含 cliPluginsExtraDirs
func (i *installer) ensureClientConfig() error {
	i.ui.Info("Ensuring Docker client config contains plugin directory...")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		i.ui.Warning("Failed to get home directory: %v", err)
		return nil // 非致命，可继续
	}

	configPath := filepath.Join(homeDir, ".docker", "config.json")
	targetDir := i.cfg.PluginDir

	// 更新 config.json
	err = utils.UpdateJSONConfigFile(configPath, func(config map[string]interface{}) error {
		extraDirs, ok := config["cliPluginsExtraDirs"].([]interface{})
		if !ok {
			config["cliPluginsExtraDirs"] = []interface{}{targetDir}
			return nil
		}
		for _, dir := range extraDirs {
			if s, ok := dir.(string); ok && s == targetDir {
				return nil
			}
		}
		config["cliPluginsExtraDirs"] = append(extraDirs, targetDir)
		return nil
	})
	if err != nil {
		i.ui.Warning("Failed to update client config %s: %v", configPath, err)
		return err
	}

	i.ui.Success("Docker client config updated successfully: %s", configPath)
	return nil
}
