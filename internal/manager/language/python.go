package language

import (
	"github.com/bookandmusic/dev-tools/internal/ui"
)

type PythonManager struct{}

func (p PythonManager) Install(version string, ui ui.UI) {
	ui.Info("Installing Python...")
}

func (p PythonManager) Uninstall(version string, ui ui.UI) {
	ui.Info("Uninstalling Python...")
}

func (p PythonManager) List(ui ui.UI) {
	ui.Info("Listing Python...")
}

func (p PythonManager) Active(ui ui.UI) {
	ui.Info("Activating Python...")
}
