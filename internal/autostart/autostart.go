// Package autostart keeps the XDG autostart entry in sync with the
// configuration.
package autostart

import (
	"fmt"
	"os"
	"path/filepath"
)

// Sync creates or removes the autostart desktop entry to match
// enabled. The Exec path always points at the current binary, so a
// moved binary heals itself on the next run.
func Sync(enabled bool) error {
	path, err := entryPath()
	if err != nil {
		return err
	}
	if !enabled {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	entry := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=laucha
Comment=Minimalist keyboard-driven launcher
Exec="%s"
Terminal=false
X-GNOME-Autostart-enabled=true
`, exe)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(entry), 0o644)
}

func entryPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "autostart", "laucha.desktop"), nil
}
