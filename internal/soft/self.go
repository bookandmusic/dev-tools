package soft

import (
	"github.com/bookandmusic/dev-tools/internal/config"
	"github.com/bookandmusic/dev-tools/internal/soft/self"
	"github.com/bookandmusic/dev-tools/internal/ui"
)

func NewSelfManager(cfg *config.GlobalConfig, ui ui.UI, InstallDir string) SoftManager {
	return &self.SelfManager{
		Config: cfg,
		UI:     ui,
		InstallParams: &self.InstallParams{
			InstallDir: InstallDir,
		},
	}
}
