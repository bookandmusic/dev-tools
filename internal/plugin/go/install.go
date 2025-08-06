package goplugin

import (
	"fmt"
	"os"
	"path/filepath"

	cli "github.com/urfave/cli/v2"
	yaml "gopkg.in/yaml.v3"

	"github.com/bookandmusic/dev-tools/internal/config"
	"github.com/bookandmusic/dev-tools/internal/ui"
	"github.com/bookandmusic/dev-tools/internal/utils"
)

type InstallPlugin struct{}

func (i *InstallPlugin) Name() string {
	return "install"
}

func (i *InstallPlugin) Description() string {
	return "Install dev-tools to a directory and setup environment variables"
}

func (i *InstallPlugin) Command() *cli.Command {
	return &cli.Command{
		Name:  "install",
		Usage: "Install dev-tools to a directory and setup environment variables",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "root-dir",
				Aliases: []string{"r"},
				Usage:   "Installation root directory",
				Value:   "~/.tools",
			},
		},
		Action: func(cCtx *cli.Context) error {
			rootDir := cCtx.String("root-dir")
			return i.install(rootDir, ui.Console)
		},
	}
}

func (i *InstallPlugin) install(rootDir string, ui ui.UI) error {
	rootAbsDir := utils.ExpandHomeAndEnv(rootDir)
	binDir := filepath.Join(rootAbsDir, "bin")
	pluginDir := filepath.Join(rootAbsDir, "plugins")

	ui.Info("Installing dev-tools to: %s", rootAbsDir)

	// 1️⃣ 创建安装目录及 bin 子目录
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		return fmt.Errorf("failed to create plugins directory: %w", err)
	}

	// 2️⃣ 生成默认 YAML 配置文件
	configFile := filepath.Join(rootAbsDir, "config.yaml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		cfg := config.Config{
			RootDir: rootAbsDir,
			BinDir:  binDir,
		}
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}
		if err := os.WriteFile(configFile, data, 0o644); err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}
		ui.Info("Generated default config file:%s", configFile)
	} else {
		ui.Warning("Config file already exists, skipped generation:%s", configFile)
	}

	// 3️⃣ 将当前执行文件复制到 bin 目录
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}
	targetPath := filepath.Join(binDir, "dtl")

	if err := utils.CopyFile(execPath, targetPath); err != nil {
		return fmt.Errorf("failed to copy executable to bin: %w", err)
	}
	ui.Info("Copied executable to:%s", targetPath)

	// 4️⃣ 复制插件文件
	// 判断当前路径是否有 plugins 目录
	if info, err := os.Stat(pluginDir); err == nil && info.IsDir() {
		utils.CopyDirWithProgress("plugins", pluginDir, ui)
	}

	// 4️⃣ 设置环境变量 & PATH（使用 bin 目录）
	err = utils.AddEnvToUserProfile(map[string]string{
		"DEV_TOOLS_HOME": rootAbsDir,
		"PATH":           binDir,
	})
	if err != nil {
		return fmt.Errorf("failed to setup environment variables: %w", err)
	}

	ui.Info("Installation complete.\nPlease restart your shell or run 'source %s' to apply changes.\n", utils.DetectProfileFile())
	return nil
}
