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
	desk.SetSystemTrayMenu(fyne.NewMenu("laucha",
		fyne.NewMenuItem(i18n.T("Show"), func() { b.show() }),
	))
	desk.SetSystemTrayIcon(fyne.NewStaticResource("tray.png", assets.TrayPNG))
	go publishTrayIconName()
	return true
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
