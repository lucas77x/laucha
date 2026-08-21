package usage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRecordAndStatsPersist(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	s, err := Open()
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := s.Record("/home/u/notas.txt"); err != nil {
		t.Fatalf("Record: %v", err)
	}
	if err := s.Record("/home/u/notas.txt"); err != nil {
		t.Fatalf("Record: %v", err)
	}
	if count, _ := s.Stats("/home/u/notas.txt"); count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	reopened, err := Open()
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer reopened.Close()
	count, last := reopened.Stats("/home/u/notas.txt")
	if count != 2 {
		t.Errorf("persisted count = %d, want 2", count)
	}
	if last.IsZero() {
		t.Error("persisted last opened is zero")
	}
	if unknown, _ := reopened.Stats("/other"); unknown != 0 {
		t.Errorf("unknown path count = %d, want 0", unknown)
	}
}

func TestDBPathFallsBackToHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")

	path, err := dbPath()
	if err != nil {
		t.Fatalf("dbPath: %v", err)
	}
	if want := filepath.Join(home, ".local", "share", "laucha", "usage.db"); path != want {
		t.Errorf("dbPath = %q, want %q", path, want)
	}
}

func TestOpenFailsOnCorruptDatabase(t *testing.T) {
	base := t.TempDir()
	t.Setenv("XDG_DATA_HOME", base)
	dir := filepath.Join(base, "laucha")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "usage.db"), []byte("this is not sqlite"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := Open(); err == nil {
		t.Error("Open on a corrupt database succeeded, want error")
	}
}

func TestOpenFailsWhenDataHomeIsAFile(t *testing.T) {
	dir := t.TempDir()
	blocked := filepath.Join(dir, "not-a-dir")
	if err := os.WriteFile(blocked, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_DATA_HOME", blocked)

	if _, err := Open(); err == nil {
		t.Error("Open with a file as data home succeeded, want error")
	}
}

func TestRecordAfterCloseReturnsError(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	s, err := Open()
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if err := s.Record("/x"); err == nil {
		t.Error("Record after Close succeeded, want error")
	}
}
