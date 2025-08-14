package adapter

import (
	"github.com/spf13/cobra"

	"github.com/bookandmusic/dev-tools/internal/config"
	"github.com/bookandmusic/dev-tools/internal/soft/ohmyzsh"
	"github.com/bookandmusic/dev-tools/internal/ui"
)

type OhMyzshPlugin struct {
	cfg     *config.GlobalConfig
	ui      ui.UI
	manager *ohmyzsh.OhMyzshManager
}

func NewOhMyzshPlugin(ui ui.UI, cfg *config.GlobalConfig) *OhMyzshPlugin {
	return &OhMyzshPlugin{
		ui:  ui,
		cfg: cfg,
	}
}

func (i *OhMyzshPlugin) Name() string {
	return "ohmyzsh"
}

func (i *OhMyzshPlugin) Description() string {
	return "manage ohmyzsh for install, uninstall, update"
}

func (i *OhMyzshPlugin) Command() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   i.Name(),
		Short: i.Description(),
		// 在任何子命令执行前，先初始化 manager
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// 初始化 manager 并保存
			i.manager = ohmyzsh.NewOhMyzshManager(i.ui, i.cfg.OhMyzsh, i.cfg.Common)
			return nil
		},
	}

	// install 子命令
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "OhMyzsh Install",
		RunE: func(cmd *cobra.Command, args []string) error {
			installDir, err := cmd.Flags().GetString("install-dir")
			if err != nil {
				return err
			}
			if !cmd.Flags().Changed("install-dir") {
				installDir = ""
			}
			return i.manager.Install(installDir)
		},
	}
	// 统一在 rootCmd 定义所有参数，这样 PersistentPreRunE 可以读取
	installCmd.PersistentFlags().String("install-dir", "~/.ohmyzsh", "Directory where the ohmzysh should be installed")
	// uninstall 子命令
	uninstallCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "OhMyzsh Uninstall",
		RunE: func(cmd *cobra.Command, args []string) error {
			return i.manager.Uninstall()
		},
	}
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "OhMyzsh Update",
		RunE: func(cmd *cobra.Command, args []string) error {
			return i.manager.Update()
		},
	}

	rootCmd.AddCommand(installCmd, uninstallCmd, updateCmd)
	return rootCmd
}
