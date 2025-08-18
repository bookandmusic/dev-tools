package script

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/bookandmusic/dev-tools/internal/ui"
	"github.com/bookandmusic/dev-tools/internal/utils"
)

type ShellExecutor struct {
	ui ui.UI
}

func NewShellExecutor(ui ui.UI) *ShellExecutor {
	return &ShellExecutor{
		ui: ui,
	}
}

// ScriptPath 生成脚本文件路径
func (s *ShellExecutor) ScriptPath(basePath string, names ...string) string {
	return filepath.Join(append([]string{basePath}, names...)...) + ".sh"
}

// Exec 执行脚本，参数来自 cobra.Command
func (s *ShellExecutor) Exec(scriptPath string, cmd *cobra.Command, args []string) error {
	// 验证脚本路径是否存在
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return s.NotFoundError(scriptPath)
	}

	// 构建命令参数
	cmdArgs := []string{scriptPath}

	// 添加flag参数
	cmd.Flags().Visit(func(f *pflag.Flag) {
		// 对flag名称和值进行基本验证
		if f.Name != "" && f.Value.String() != "" {
			cmdArgs = append(cmdArgs, "--"+f.Name, f.Value.String())
		}
	})

	// 添加位置参数
	cmdArgs = append(cmdArgs, args...)
	ctx := context.Background()
	// 使用统一的命令执行方式
	return utils.RunCommand(ctx, s.ui, nil, "bash", cmdArgs...)
}

func (s *ShellExecutor) NotFoundError(path string) error {
	return fmt.Errorf("script not found: %s", path)
}
