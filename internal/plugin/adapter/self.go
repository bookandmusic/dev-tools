package adapter

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/bookandmusic/dev-tools/internal/config"
	"github.com/bookandmusic/dev-tools/internal/manager/soft"
	"github.com/bookandmusic/dev-tools/internal/ui"
)

type InstallPlugin struct {
	cfg     *config.GlobalConfig
	ui      ui.UI
	manager soft.SoftManage
}

func NewSelfPlugin(ui ui.UI, cfg *config.GlobalConfig) *InstallPlugin {
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
			// 初始化 manager 并保存
			manage, err := soft.GetManager("self")
			if err != nil {
				return err
			}
			i.manager = manage
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
			installDir, err := cmd.Flags().GetString("install-dir")
			if err != nil {
				return err
			}
			cfg := i.cfg
			if cmd.Flags().Changed("install-dir") {
				config.NewManager().UpdateRootDir(cfg, installDir)
			}
			ctx := context.Background()
			ctx = context.WithValue(ctx, soft.ContextKey("cfg"), cfg)
			ctx = context.WithValue(ctx, soft.ContextKey("ui"), i.ui)
			return i.manager.Install(ctx)
		},
	}
	// 统一在 rootCmd 定义所有参数，这样 PersistentPreRunE 可以读取
	installCmd.PersistentFlags().String("install-dir", "~/.tools", "Directory where the tool should be installed")
	// uninstall 子命令
	uninstallCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall self",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			ctx = context.WithValue(ctx, soft.ContextKey("cfg"), i.cfg)
			ctx = context.WithValue(ctx, soft.ContextKey("ui"), i.ui)
			return i.manager.Uninstall(ctx)
		},
	}

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update self",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			ctx = context.WithValue(ctx, soft.ContextKey("cfg"), i.cfg)
			ctx = context.WithValue(ctx, soft.ContextKey("ui"), i.ui)
			return i.manager.Update(ctx)
		},
	}

	rootCmd.AddCommand(installCmd, uninstallCmd, updateCmd)
	return rootCmd
}
