package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"

	"github.com/lucas77x/laucha/assets"
	"github.com/lucas77x/laucha/internal/i18n"
)

// setupTray installs the system-tray icon and menu when the config
// enables it. Fyne appends its own Quit item to the menu.
func (b *Bar) setupTray() bool {
	if !b.cfg.Behavior.ShowTrayIcon {
		return false
	}
	desk, ok := b.app.(desktop.App)
	if !ok {
		return false
	}
	desk.SetSystemTrayMenu(fyne.NewMenu("laucha",
		fyne.NewMenuItem(i18n.T("Show"), func() { b.show() }),
	))
	desk.SetSystemTrayIcon(fyne.NewStaticResource("icon.svg", assets.IconSVG))
	return true
}
