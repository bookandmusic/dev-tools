package adapter

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/bookandmusic/dev-tools/internal/config"
	"github.com/bookandmusic/dev-tools/internal/manager/soft"
	"github.com/bookandmusic/dev-tools/internal/ui"
)

type DockerPlugin struct {
	cfg     *config.GlobalConfig
	ui      ui.UI
	manager soft.SoftManage
}

func NewDockerPlugin(ui ui.UI, cfg *config.GlobalConfig) *DockerPlugin {
	return &DockerPlugin{
		ui:  ui,
		cfg: cfg,
	}
}

func (i *DockerPlugin) Name() string {
	return "docker"
}

func (i *DockerPlugin) Description() string {
	return "manage docker for install, uninstall, update"
}

func (i *DockerPlugin) Command() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   i.Name(),
		Short: i.Description(),
		// 在任何子命令执行前，先初始化 manager
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// 初始化 manager 并保存
			manage, err := soft.GetManager("docker")
			if err != nil {
				return err
			}
			i.manager = manage
			return nil
		},
	}

	// install 子命令
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Docker Install",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			ctx = context.WithValue(ctx, soft.ContextKey("cfg"), i.cfg.Docker)
			ctx = context.WithValue(ctx, soft.ContextKey("global"), i.cfg.Common)
			ctx = context.WithValue(ctx, soft.ContextKey("ui"), i.ui)

			return i.manager.Install(ctx)
		},
	}
	// uninstall 子命令
	uninstallCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Docker Uninstall",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			ctx = context.WithValue(ctx, soft.ContextKey("cfg"), i.cfg.Docker)
			ctx = context.WithValue(ctx, soft.ContextKey("global"), i.cfg.Common)
			ctx = context.WithValue(ctx, soft.ContextKey("ui"), i.ui)
			return i.manager.Uninstall(ctx)
		},
	}
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Docker Update",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			ctx = context.WithValue(ctx, soft.ContextKey("cfg"), i.cfg.Docker)
			ctx = context.WithValue(ctx, soft.ContextKey("global"), i.cfg.Common)
			ctx = context.WithValue(ctx, soft.ContextKey("ui"), i.ui)
			ctx = context.WithValue(ctx, soft.ContextKey("env"), map[string]string{})
			return i.manager.Update(ctx)
		},
	}

	rootCmd.AddCommand(installCmd, uninstallCmd, updateCmd)
	return rootCmd
}
