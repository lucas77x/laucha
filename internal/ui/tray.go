package ui

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/systray"

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
	b.trayToggle = fyne.NewMenuItem(i18n.T("Show"), b.toggle)
	b.trayMenu = fyne.NewMenu("laucha",
		b.trayToggle,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem(i18n.T("Settings"), b.showSettings),
		fyne.NewMenuItem(i18n.T("About"), b.showAbout),
	)
	desk.SetSystemTrayMenu(b.trayMenu)
	desk.SetSystemTrayIcon(fyne.NewStaticResource("tray.png", assets.TrayPNG))
	go publishTrayIconName()
	return true
}

// setTrayVisible shows or hides the tray icon immediately: first
// activation builds the tray, later toggles flip the SNI status.
func (b *Bar) setTrayVisible(on bool) {
	if on {
		if b.trayActive {
			systray.SetStatus("Active")
		} else {
			b.trayActive = b.setupTray()
		}
	} else if b.trayActive {
		systray.SetStatus("Passive")
	}
	b.resident = b.trayActive || b.hotkeyActive
}

// refreshTrayToggle keeps the tray item in sync with visibility:
// "Show" while hidden, "Hide" while visible.
func (b *Bar) refreshTrayToggle() {
	if b.trayToggle == nil {
		return
	}
	if b.visible {
		b.trayToggle.Label = i18n.T("Hide")
	} else {
		b.trayToggle.Label = i18n.T("Show")
	}
	b.trayMenu.Refresh()
}

// publishTrayIconName works around tray hosts (ayatana indicators on
// MATE and XFCE) that ignore IconPixmap: it writes the icon to disk
// and republishes it by file path once the tray is registered.
func publishTrayIconName() {
	iconPath, err := writeTrayIcon()
	if err != nil {
		log.Printf("tray: writing icon: %v", err)
		return
	}
	for attempt := 0; attempt < 50; attempt++ {
		if systray.SetIconName(iconPath) {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	log.Printf("tray: icon name never published; tray host missing?")
}

func writeTrayIcon() (string, error) {
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".local", "share")
	}
	dir := filepath.Join(base, "laucha")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	iconPath := filepath.Join(dir, "tray.png")
	if err := os.WriteFile(iconPath, assets.TrayPNG, 0o644); err != nil {
		return "", err
	}
	return iconPath, nil
}
