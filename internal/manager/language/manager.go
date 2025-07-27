package language

import "github.com/bookandmusic/dev-tools/internal/ui"

type Manager interface {
	Install(version string, ui ui.UI)
	Uninstall(version string, ui ui.UI)
	List(ui ui.UI)
	Active(version string, ui ui.UI)
}
