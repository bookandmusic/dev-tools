package plugin

import (
	"github.com/spf13/cobra"

	"github.com/bookandmusic/dev-tools/internal/config"
	"github.com/bookandmusic/dev-tools/internal/ui"
)

var (
	registered = []*cobra.Command{}
)

func Register(cmd *cobra.Command) {
	registered = append(registered, cmd)
}

func Commands(ui ui.UI, cfg *config.GlobalConfig, rootDir string) []*cobra.Command {
	return registered
}
