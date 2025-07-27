package self

import (
	"github.com/bookandmusic/dev-tools/internal/config"
	"github.com/bookandmusic/dev-tools/internal/ui"
)

type BaseParams struct {
	UI  ui.UI                `ctx:"ui"`
	Cfg *config.GlobalConfig `ctx:"cfg"`
	Env map[string]string    `ctx:"env"`
}
