package ui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsExecutable(t *testing.T) {
	dir := t.TempDir()

	executable := filepath.Join(dir, "run.AppImage")
	if err := os.WriteFile(executable, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	plain := filepath.Join(dir, "notes.txt")
	if err := os.WriteFile(plain, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if !isExecutable(executable) {
		t.Error("isExecutable(+x file) = false, want true")
	}
	if isExecutable(plain) {
		t.Error("isExecutable(0644 file) = true, want false")
	}
	if isExecutable(dir) {
		t.Error("isExecutable(directory) = true, want false")
	}
	if isExecutable(filepath.Join(dir, "missing")) {
		t.Error("isExecutable(missing) = true, want false")
	}
}
