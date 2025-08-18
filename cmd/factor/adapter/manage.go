package adapter

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/bookandmusic/dev-tools/cmd/plugin"
	"github.com/bookandmusic/dev-tools/internal/config"
	"github.com/bookandmusic/dev-tools/internal/manager/soft"
	"github.com/bookandmusic/dev-tools/internal/ui"
)

func NewSelfPlugin(ui ui.UI, cfg *config.GlobalConfig) *cobra.Command {
	return BuildPlugin(ui, cfg, PluginSpec{
		Name:        "self",
		Description: "manage dev-tools for install, uninstall",
		ManagerName: "self",
		Config:      cfg,
		ContextMap:  nil,
		Subcommands: []SubcommandSpec{
			{
				Name:  "install",
				Short: "Install self",
				Action: func(ctx context.Context, m soft.SoftManage) error {
					cmd := ctx.Value(soft.ContextKey("cmd")).(*cobra.Command)
					installDir, _ := cmd.Flags().GetString("install-dir")
					if cmd.Flags().Changed("install-dir") {
						config.NewManager().UpdateRootDir(cfg, installDir)
					}
					return m.Install(ctx)
				},
				Flags: func(cmd *cobra.Command) {
					cmd.PersistentFlags().String("install-dir", "~/.tools", "Directory where the tool should be installed")
				},
			},
			{Name: "uninstall", Short: "Uninstall self", Action: func(ctx context.Context, m soft.SoftManage) error { return m.Uninstall(ctx) }, Flags: nil},
			{Name: "update", Short: "Update self", Action: func(ctx context.Context, m soft.SoftManage) error { return m.Update(ctx) }, Flags: nil},
		},
	})
}

func NewOhMyzshPlugin(ui ui.UI, cfg *config.GlobalConfig) *cobra.Command {
	return BuildPlugin(ui, cfg, PluginSpec{
		Name:        "ohmyzsh",
		Description: "manage ohmyzsh for install, uninstall, update",
		ManagerName: "ohmyzsh",
		Config:      cfg.OhMyzsh,
		ContextMap:  map[soft.ContextKey]any{"env": map[string]string{}}, // 自定义 env
		Subcommands: createStandardSubcommands("OhMyzsh"),
	})
}

func NewDockerPlugin(ui ui.UI, cfg *config.GlobalConfig) *cobra.Command {
	return BuildPlugin(ui, cfg, PluginSpec{
		Name:        "docker",
		Description: "managedocker for install, uninstall",
		ManagerName: "docker",
		Config:      cfg,
		ContextMap:  nil,
		Subcommands: createStandardSubcommands("Docker"),
	})
}

// createStandardSubcommands 创建标准的子命令（install, uninstall, update）
func createStandardSubcommands(prefix string) []SubcommandSpec {
	return []SubcommandSpec{
		{Name: "install", Short: prefix + " Install", Action: func(ctx context.Context, m soft.SoftManage) error { return m.Install(ctx) }, Flags: nil},
		{Name: "uninstall", Short: prefix + " Uninstall", Action: func(ctx context.Context, m soft.SoftManage) error { return m.Uninstall(ctx) }, Flags: nil},
		{Name: "update", Short: prefix + " Update", Action: func(ctx context.Context, m soft.SoftManage) error { return m.Update(ctx) }, Flags: nil},
	}
}

func LoadPluginsFromAdapter(ui ui.UI, cfg *config.GlobalConfig) {
	plugin.Register(NewDockerPlugin(ui, cfg))
	plugin.Register(NewOhMyzshPlugin(ui, cfg))
	plugin.Register(NewSelfPlugin(ui, cfg))
}