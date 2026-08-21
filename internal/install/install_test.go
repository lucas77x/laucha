package install

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallAndUninstall(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	if err := Install(); err != nil {
		t.Fatalf("Install: %v", err)
	}

	entry := filepath.Join(dataHome(), "applications", "laucha.desktop")
	data, err := os.ReadFile(entry)
	if err != nil {
		t.Fatalf("desktop entry not written: %v", err)
	}
	exe, _ := os.Executable()
	if !strings.Contains(string(data), "Exec=\""+exe+"\"") {
		t.Errorf("entry missing quoted Exec:\n%s", data)
	}
	icon := filepath.Join(dataHome(), "icons", "hicolor", "scalable", "apps", "laucha.svg")
	if _, err := os.Stat(icon); err != nil {
		t.Errorf("icon not written: %v", err)
	}

	if err := Uninstall(); err != nil {
		t.Fatalf("Uninstall: %v", err)
	}
	if _, err := os.Stat(entry); !os.IsNotExist(err) {
		t.Error("desktop entry still present after Uninstall")
	}
	if err := Uninstall(); err != nil {
		t.Errorf("second Uninstall must be a no-op, got %v", err)
	}
}
