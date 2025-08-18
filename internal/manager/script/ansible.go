package script

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/apenella/go-ansible/v2/pkg/playbook"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/bookandmusic/dev-tools/internal/ui"
)

type AnsibleExecutor struct {
	ui ui.UI
}

func NewAnsibleExecutor(ui ui.UI) *AnsibleExecutor {
	return &AnsibleExecutor{ui: ui}
}

func (a *AnsibleExecutor) ScriptPath(basePath string, names ...string) string {
	return filepath.Join(append([]string{basePath}, names...)...) + ".yml"
}

func (a *AnsibleExecutor) Exec(scriptPath string, cmd *cobra.Command, args []string) error {
	ansiblePlaybookOptions := &playbook.AnsiblePlaybookOptions{
		Connection: "local",
		Inventory:  "127.0.0.1,",     // 逗号结尾是必须的
		ExtraVars:  map[string]any{}, // 自定义变量
	}

	// 遍历所有已设置的 flags
	cmd.Flags().Visit(func(f *pflag.Flag) {
		ansiblePlaybookOptions.ExtraVars[f.Name] = f.Value.String()
	})

	// 这里你可以选择是否用 args 传给 ansible，有需要的话可以自行追加

	// 执行 playbook
	err := playbook.NewAnsiblePlaybookExecute(scriptPath).
		WithPlaybookOptions(ansiblePlaybookOptions).
		Execute(context.TODO())

	return err
}

func (a *AnsibleExecutor) NotFoundError(path string) error {
	return fmt.Errorf("playbook not found: %s", path)
}
