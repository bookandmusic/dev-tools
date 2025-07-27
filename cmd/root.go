package cmd

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/bookandmusic/dev-tools/internal/config"
	"github.com/bookandmusic/dev-tools/internal/plugin"
	"github.com/bookandmusic/dev-tools/internal/ui"
)

var (
	rootDir        string
	configFile     string
	userName       string
	debug          bool
	rootDirChanged bool
	configChanged  bool

	rootCmd = &cobra.Command{
		Use:   "dev-tools",
		Short: "A professional Go application",
	}
)

func init() {
	// 定义全局 flag
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug mode")
	rootCmd.PersistentFlags().StringVarP(&rootDir, "root-dir", "r", ".", "Installation root directory")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config.yaml", "Load configuration from FILE")
	// 根据预解析的 rootDir 先加载插件注册命令
	ui := ui.NewConsoleUI(debug)
	// 预解析全局 flags（必须在命令注册前调用，否则 cobra 会报错）
	// 这里用 rootCmd.ParseFlags 解析 os.Args ，但忽略错误（比如 -h 会出错）
	if err := rootCmd.ParseFlags(os.Args[1:]); err != nil {
		ui.Debug("parse flags error: %s", err)
	}

	// 保存 changed 状态
	rootDirChanged = rootCmd.PersistentFlags().Changed("root-dir")
	configChanged = rootCmd.PersistentFlags().Changed("config")

	if !rootDirChanged {
		if envHomeDir := os.Getenv("DEV_TOOLS_HOME"); envHomeDir != "" {
			ui.Debug("use env DEV_TOOLS_HOME: %s", envHomeDir)
			rootDir = envHomeDir
		}
	}
	ui.Debug("use root dir: %s", rootDir)

	if !configChanged {
		ui.Debug("use default config file: %s", configFile)
		configFile = filepath.Join(rootDir, configFile)
	}
	ui.Debug("use config file: %s", configFile)
	currentUser, err := user.Current()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	userName = currentUser.Username

	cfg, err := config.LoadConfigFile(
		debug,
		configFile,
	)
	if err != nil {
		cfg = config.GenerateDefaultConfig(rootDir)
		ui.Debug("generate default config")
	}
	plugin.LoadPluginsToRefistry(cfg.Common.RootDir, ui, cfg)
	plugin.RegistryPlugins()
	plugins := plugin.GetRegisteredPlugins()
	for _, p := range plugins {
		rootCmd.AddCommand(p)
	}
}

// Run 执行入口
func Execute() error {
	return rootCmd.Execute()
}
