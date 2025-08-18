package script

import (
	"github.com/spf13/cobra"
)

type PluginExecutor interface {
	ScriptPath(basePath string, names ...string) string
	Exec(scriptPath string, cmd *cobra.Command, args []string) error
	NotFoundError(path string) error
}
