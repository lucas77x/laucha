package index

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/lucas77x/laucha/internal/config"
)

func TestReconfigureSwapsRootsAndFilter(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	rootA := t.TempDir()
	rootB := t.TempDir()
	mustWrite(t, filepath.Join(rootA, "a.txt"))
	mustWrite(t, filepath.Join(rootB, "b.txt"))

	idx, err := New([]string{rootA}, config.Filter{Mode: "exclude"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer idx.Close()

	waitFor(t, "a.txt indexed", func() bool { return hasName(idx, "a.txt") })

	idx.Reconfigure([]string{rootB}, config.Filter{Mode: "exclude"})

	waitFor(t, "index swapped to rootB", func() bool {
		return hasName(idx, "b.txt") && !hasName(idx, "a.txt")
	})
}

func hasName(idx *Index, name string) bool {
	for _, e := range idx.Entries() {
		if e.Name == name {
			return true
		}
	}
	return false
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}
