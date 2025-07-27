package adapter

import (
	"github.com/spf13/cobra"

	"github.com/bookandmusic/dev-tools/internal/config"
	"github.com/bookandmusic/dev-tools/internal/soft"
	"github.com/bookandmusic/dev-tools/internal/ui"
)

type InstallPlugin struct {
	cfg     *config.GlobalConfig
	ui      ui.UI
	manager soft.SoftManager
}

func NewInstallPlugin(ui ui.UI, cfg *config.GlobalConfig) *InstallPlugin {
	return &InstallPlugin{
		ui:  ui,
		cfg: cfg,
	}
}

func (i *InstallPlugin) Name() string {
	return "self"
}

func (i *InstallPlugin) Description() string {
	return "manage dev-tools for install, uninstall"
}

func (i *InstallPlugin) Command() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   i.Name(),
		Short: i.Description(),
		// 在任何子命令执行前，先初始化 manager
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			installDir, err := cmd.Flags().GetString("install-dir")
			if err != nil {
				return err
			}
			// 这里解析其它参数也可以

			// 初始化 manager 并保存
			i.manager = soft.NewSelfManager(i.cfg, i.ui, installDir)
			return nil
		},
	}

	// 可继续定义其他公共参数

	// install 子命令
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install self",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 这里直接用已经初始化好的 manager
			return i.manager.Install()
		},
	}
	// 统一在 rootCmd 定义所有参数，这样 PersistentPreRunE 可以读取
	installCmd.PersistentFlags().String("install-dir", "~/.tools", "Directory where the tool should be installed")
	// uninstall 子命令
	uninstallCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall self",
		RunE: func(cmd *cobra.Command, args []string) error {
			return i.manager.Uninstall()
		},
	}

	rootCmd.AddCommand(installCmd, uninstallCmd)
	return rootCmd
}
