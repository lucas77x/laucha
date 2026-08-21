package index

import (
	"os"
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

func TestWatcherKeepsIndexLive(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "seed.txt"))

	idx, err := New([]string{root}, config.Filter{Mode: "exclude", Extensions: []string{".log"}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer idx.Close()
	waitFor(t, "initial walk", func() bool { return hasName(idx, "seed.txt") })

	// A file created after startup must appear without any rescan.
	mustWrite(t, filepath.Join(root, "fresh.txt"))
	waitFor(t, "created file indexed", func() bool { return hasName(idx, "fresh.txt") })

	// Filtered files must not appear even when the watcher sees them.
	mustWrite(t, filepath.Join(root, "noise.log"))
	mustWrite(t, filepath.Join(root, "after-noise.txt"))
	waitFor(t, "marker after noise", func() bool { return hasName(idx, "after-noise.txt") })
	if hasName(idx, "noise.log") {
		t.Error("filtered .log file leaked into the live index")
	}

	// New directories are walked and watched recursively.
	mustWrite(t, filepath.Join(root, "sub", "nested.txt"))
	waitFor(t, "nested file indexed", func() bool { return hasName(idx, "nested.txt") })

	// Removals leave the index.
	if err := os.Remove(filepath.Join(root, "fresh.txt")); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "removed file dropped", func() bool { return !hasName(idx, "fresh.txt") })
}

func TestRecentOrdersNewestFirst(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	root := t.TempDir()
	older := filepath.Join(root, "older.txt")
	newer := filepath.Join(root, "newer.txt")
	mustWrite(t, older)
	mustWrite(t, newer)
	past := time.Now().Add(-time.Hour)
	if err := os.Chtimes(older, past, past); err != nil {
		t.Fatal(err)
	}

	idx, err := New([]string{root}, config.Filter{Mode: "exclude"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer idx.Close()
	waitFor(t, "walk", func() bool { return len(idx.Entries()) == 2 })

	recent := idx.Recent(1)
	if len(recent) != 1 || recent[0].Name != "newer.txt" {
		t.Errorf("Recent(1) = %v, want newer.txt first", recent)
	}
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
