package loader

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/bookandmusic/dev-tools/internal/manager/script"
)

func CreateCommandTree(basePath string, meta *PluginMeta, executor script.PluginExecutor) *cobra.Command {
	root := &cobra.Command{
		Use:   meta.Name,
		Short: meta.Description,
	}

	if len(meta.Commands) > 0 {
		for cmdName, cmdDef := range meta.Commands {
			sub := buildCommand(basePath, []string{cmdName}, cmdDef, executor)
			root.AddCommand(sub)
		}
	} else {
		path := executor.ScriptPath(basePath, meta.Name)
		root.RunE = makeRunE(path, executor)
	}

	return root
}

func buildCommand(basePath string, pathParts []string, cmdDef Command, executor script.PluginExecutor) *cobra.Command {
	cmdName := pathParts[len(pathParts)-1]
	cmd := &cobra.Command{
		Use:   cmdName,
		Short: cmdDef.Usage,
	}

	// 添加 flags
	for _, opt := range cmdDef.Options {
		cmd.Flags().StringP(opt.Name, opt.Short, opt.Value, opt.Description)
	}

	scriptPath := executor.ScriptPath(basePath, pathParts...)
	cmd.RunE = makeRunE(scriptPath, executor)

	for subName, subCmdDef := range cmdDef.Subcommands {
		newPath := append(pathParts, subName)
		subCmd := buildCommand(basePath, newPath, subCmdDef, executor)
		cmd.AddCommand(subCmd)
	}

	return cmd
}

func makeRunE(scriptPath string, executor script.PluginExecutor) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			return executor.NotFoundError(scriptPath)
		}
		return executor.Exec(scriptPath, cmd, args)
	}
}
