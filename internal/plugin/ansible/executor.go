package ansible

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/apenella/go-ansible/v2/pkg/playbook"
	cli "github.com/urfave/cli/v2"
)

type AnsibleExecutor struct{}

func (a *AnsibleExecutor) ScriptPath(basePath string, names ...string) string {
	return filepath.Join(append([]string{basePath}, names...)...) + ".yml"
}

func (a *AnsibleExecutor) Exec(scriptPath string, c *cli.Context) error {
	// 创建选项
	ansiblePlaybookOptions := &playbook.AnsiblePlaybookOptions{
		Connection: "local",
		Inventory:  "127.0.0.1,",     // 逗号结尾是必须的（Ansible 要求）
		ExtraVars:  map[string]any{}, // 添加自定义变量支持
	}

	// 遍历所有设置过的 flags
	for _, flag := range c.Command.Flags {
		names := flag.Names()
		mainName := names[0]

		var hasUserInput bool

		// 优先检查用户是否传入任何别名或主名
		for _, alias := range names {
			if c.IsSet(alias) {
				val := c.Value(mainName)
				if val != nil {
					ansiblePlaybookOptions.ExtraVars[mainName] = fmt.Sprintf("%v", val)
				}
				hasUserInput = true
				break
			}
		}

		// 如果用户没传入，使用默认值（如果有）
		if !hasUserInput {
			val := c.Value(mainName)
			if val != nil {
				ansiblePlaybookOptions.ExtraVars[mainName] = fmt.Sprintf("%v", val)
			}
		}
	}

	// 执行 playbook
	err := playbook.NewAnsiblePlaybookExecute(scriptPath).
		WithPlaybookOptions(ansiblePlaybookOptions).
		Execute(context.TODO())

	return err
}

func (a *AnsibleExecutor) NotFoundError(path string) error {
	return fmt.Errorf("playbook not found: %s", path)
}
