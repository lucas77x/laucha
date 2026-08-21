package autostart

import (
	"os"
	"path/filepath"
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
	if !strings.Contains(string(data), "Exec=\""+exe+"\"") {
		t.Errorf("entry missing quoted Exec for %s:\n%s", exe, data)
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

func TestSyncTrueFailsWhenConfigDirIsAFile(t *testing.T) {
	dir := t.TempDir()
	blocked := filepath.Join(dir, "not-a-dir")
	if err := os.WriteFile(blocked, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", blocked)

	if err := Sync(true); err == nil {
		t.Error("Sync(true) with a file as config home succeeded, want error")
	}
}

func TestSyncFalseFailsWhenEntryIsANonEmptyDir(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	path, err := entryPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(path, "child"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := Sync(false); err == nil {
		t.Error("Sync(false) with a populated dir at the entry path succeeded, want error")
	}
}
