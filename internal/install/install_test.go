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

func TestDataHomeFallsBackToHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")

	if got, want := dataHome(), filepath.Join(home, ".local", "share"); got != want {
		t.Errorf("dataHome = %q, want %q", got, want)
	}
}

func TestInstallFailsWhenDataHomeIsAFile(t *testing.T) {
	dir := t.TempDir()
	blocked := filepath.Join(dir, "not-a-dir")
	if err := os.WriteFile(blocked, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_DATA_HOME", blocked)

	if err := Install(); err == nil {
		t.Error("Install with a file as data home succeeded, want error")
	}
}

func TestUninstallFailsOnNonEmptyDirEntry(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	entry := filepath.Join(dataHome(), "applications", "laucha.desktop")
	if err := os.MkdirAll(filepath.Join(entry, "child"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := Uninstall(); err == nil {
		t.Error("Uninstall with a populated dir at the entry path succeeded, want error")
	}
}
