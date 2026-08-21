// Package install integrates laucha with the desktop: an application
// menu entry and icon, so the launcher shows up like any other app.
package install

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lucas77x/laucha/assets"
)

// Install writes the menu entry and icon pointing at the current
// binary.
func Install() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	iconDir := filepath.Join(dataHome(), "icons", "hicolor", "scalable", "apps")
	if err := os.MkdirAll(iconDir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(iconDir, "laucha.svg"), assets.IconSVG, 0o644); err != nil {
		return err
	}

	appsDir := filepath.Join(dataHome(), "applications")
	if err := os.MkdirAll(appsDir, 0o755); err != nil {
		return err
	}
	entry := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=laucha
Comment=Minimalist keyboard-driven launcher
Exec="%s"
Icon=laucha
Terminal=false
Categories=Utility;
StartupNotify=false
`, exe)
	return os.WriteFile(filepath.Join(appsDir, "laucha.desktop"), []byte(entry), 0o644)
}

// Uninstall removes the menu entry and icon.
func Uninstall() error {
	paths := []string{
		filepath.Join(dataHome(), "applications", "laucha.desktop"),
		filepath.Join(dataHome(), "icons", "hicolor", "scalable", "apps", "laucha.svg"),
	}
	for _, path := range paths {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func dataHome() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return xdg
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share")
}
