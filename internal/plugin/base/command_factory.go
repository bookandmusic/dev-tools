package base

import (
	"os"

	cli "github.com/urfave/cli/v2"
)

func CreateCommandTree(basePath string, meta PluginMeta, executor PluginExecutor) *cli.Command {
	root := &cli.Command{
		Name:  meta.Name,
		Usage: meta.Description,
	}

	if len(meta.Commands) > 0 {
		for cmdName, cmdDef := range meta.Commands {
			sub := buildCommand(basePath, []string{cmdName}, cmdDef, executor)
			root.Subcommands = append(root.Subcommands, sub)
		}
	} else {
		path := executor.ScriptPath(basePath, meta.Name)
		root.Action = makeAction(path, executor)
	}

	return root
}

func buildCommand(basePath string, pathParts []string, cmdDef Command, executor PluginExecutor) *cli.Command {
	cmd := &cli.Command{
		Name:  pathParts[len(pathParts)-1],
		Usage: cmdDef.Usage,
		Flags: convertOptions(cmdDef.Options),
	}

	scriptPath := executor.ScriptPath(basePath, pathParts...)
	cmd.Action = makeAction(scriptPath, executor)

	for subName, subCmd := range cmdDef.Subcommands {
		newPath := append(pathParts, subName)
		sub := buildCommand(basePath, newPath, subCmd, executor)
		cmd.Subcommands = append(cmd.Subcommands, sub)
	}

	return cmd
}

func makeAction(path string, executor PluginExecutor) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return executor.NotFoundError(path)
		}
		return executor.Exec(path, c)
	}
}

func convertOptions(opts []Option) []cli.Flag {
	var flags []cli.Flag
	for _, o := range opts {
		flags = append(flags, &cli.StringFlag{
			Name:    o.Name,
			Aliases: []string{o.Short},
			Usage:   o.Description,
			Value:   o.Value,
		})
	}
	return flags
}
