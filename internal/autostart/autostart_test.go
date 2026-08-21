package autostart

import (
	"os"
	"strings"
	"testing"
)

func TestSyncCreatesAndRemovesEntry(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	if err := Sync(true); err != nil {
		t.Fatalf("Sync(true): %v", err)
	}
	path, err := entryPath()
	if err != nil {
		t.Fatalf("entryPath: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("entry not written: %v", err)
	}
	exe, _ := os.Executable()
	if !strings.Contains(string(data), "Exec="+exe) {
		t.Errorf("entry missing Exec=%s:\n%s", exe, data)
	}

	if err := Sync(false); err != nil {
		t.Fatalf("Sync(false): %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("entry still exists after Sync(false)")
	}
}

func TestSyncFalseWithoutEntryIsNoop(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	if err := Sync(false); err != nil {
		t.Errorf("Sync(false) without entry: %v", err)
	}
}
