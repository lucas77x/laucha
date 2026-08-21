package usage

import (
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
