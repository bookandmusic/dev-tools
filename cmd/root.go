package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/bookandmusic/dev-tools/cmd/factor/adapter"
	"github.com/bookandmusic/dev-tools/cmd/factor/loader"
	"github.com/bookandmusic/dev-tools/cmd/plugin"
	"github.com/bookandmusic/dev-tools/internal/config"
	"github.com/bookandmusic/dev-tools/internal/ui"
	"github.com/bookandmusic/dev-tools/internal/utils"
)

var (
	rootDir          string
	rootDirChange    bool
	configFile       string
	configFileChange bool
	debug            bool

	rootCmd = &cobra.Command{
		Use:   "dev-tools",
		Short: "A professional Go application",
	}
)

func init() {
	// 定义全局 flag
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug mode")
	rootCmd.PersistentFlags().StringVarP(&rootDir, "root-dir", "r", "~/.tools", "tools root directory")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config.yml", "Load configuration from FILE")
	// 预解析全局 flags（必须在命令注册前调用，否则 cobra 会报错）
	// 这里用 rootCmd.ParseFlags 解析 os.Args ，但忽略错误（比如 -h 会出错）
	_ = rootCmd.ParseFlags(os.Args[1:])

	// 根据预解析的 rootDir 先加载插件注册命令
	ui := ui.NewConsoleUI(debug)
	rootDirChange = rootCmd.Flags().Changed("root-dir")
	configFileChange = rootCmd.Flags().Changed("config")

	// 初始化配置管理器
	cfgMgr := config.NewManager()

	// 加载配置
	cfg, err := cfgMgr.LoadConfigWithFallback(
		configFile,
		rootDir,
		configFileChange,
		rootDirChange,
	)
	if err != nil {
		ui.Debug("Failed to load config: %v", err)
	}
	if cfg == nil {
		cfg = &config.GlobalConfig{Common: &config.CommonConfig{}}
	}
	workdir := utils.ExpandAbsDir(cfgMgr.DetermineWorkDir(cfg.Common.RootDir, rootDir, rootDirChange))
	cfg.Common.WorkDir = workdir
	cfg.Common.Debug = debug
	rootPath := utils.ExpandAbsDir(cfgMgr.DetermineRootDir(cfg.Common.RootDir, rootDir, rootDirChange))
	cfgMgr.SetDefaults(cfg, rootPath)
	adapter.LoadPluginsFromAdapter(ui, cfg)
	loader.LoadPluginsFromLoader(rootDir, ui)
	cmds := plugin.Commands(ui, cfg, workdir)
	for _, p := range cmds {
		rootCmd.AddCommand(p)
	}
}

// Run 执行入口
func Execute() error {
	return rootCmd.Execute()
}
