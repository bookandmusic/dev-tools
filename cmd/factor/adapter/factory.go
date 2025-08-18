package adapter

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/bookandmusic/dev-tools/internal/config"
	"github.com/bookandmusic/dev-tools/internal/manager/soft"
	"github.com/bookandmusic/dev-tools/internal/ui"
)

type SubcommandSpec struct {
	Name   string
	Short  string
	Action func(ctx context.Context, manager soft.SoftManage) error
	Flags  func(cmd *cobra.Command) // 可选：不同命令需要的 flags
}

type PluginSpec struct {
	Name        string
	Description string
	ManagerName string
	Config      any
	Subcommands []SubcommandSpec
	ContextMap  map[soft.ContextKey]any // 新增：自定义 context
}

func BuildPlugin(ui ui.UI, cfg *config.GlobalConfig, spec PluginSpec) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   spec.Name,
		Short: spec.Description,
	}

	var manager soft.SoftManage

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		m, err := soft.GetManager(spec.ManagerName)
		if err != nil {
			return err
		}
		manager = m
		return nil
	}

	for _, sub := range spec.Subcommands {
		subCmd := &cobra.Command{
			Use:   sub.Name,
			Short: sub.Short,
			RunE: func(cmd *cobra.Command, args []string) error {
				ctx := context.Background()
				// 默认公共参数
				ctx = context.WithValue(ctx, soft.ContextKey("cfg"), spec.Config)
				ctx = context.WithValue(ctx, soft.ContextKey("global"), cfg.Common)
				ctx = context.WithValue(ctx, soft.ContextKey("ui"), ui)

				// 插件自定义 context
				for k, v := range spec.ContextMap {
					ctx = context.WithValue(ctx, k, v)
				}

				// 有时需要把 cmd 本身传进去，方便读取 flags
				ctx = context.WithValue(ctx, soft.ContextKey("cmd"), cmd)

				return sub.Action(ctx, manager)
			},
		}
		if sub.Flags != nil {
			sub.Flags(subCmd)
		}
		rootCmd.AddCommand(subCmd)
	}

	return rootCmd
}
