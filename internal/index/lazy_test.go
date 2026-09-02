package index

import (
	"path/filepath"
	"testing"

	"github.com/lucas77x/laucha/internal/config"
)

func TestLazyOpensOnlyWhenAsked(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "a.txt"))

	l := NewLazy([]string{root}, config.Filter{Mode: "exclude"})
	defer l.Close()

	if l.Opened() {
		t.Fatal("the index opened before anything was asked of it")
	}

	l.Entries() // the bar reaches the provider only with file search on

	if !l.Opened() {
		t.Fatal("asking for entries did not open the index")
	}
	waitFor(t, "a.txt indexed", func() bool { return hasName(l.idx, "a.txt") })
}

func TestLazyRemembersSettingsChangedBeforeOpening(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	first := t.TempDir()
	second := t.TempDir()
	mustWrite(t, filepath.Join(first, "old.txt"))
	mustWrite(t, filepath.Join(second, "new.txt"))

	l := NewLazy([]string{first}, config.Filter{Mode: "exclude"})
	defer l.Close()

	l.Reconfigure([]string{second}, config.Filter{Mode: "exclude"})
	if l.Opened() {
		t.Fatal("reconfiguring opened an index nobody asked for")
	}

	l.Recent(10)

	waitFor(t, "the index opened on the new root", func() bool {
		return hasName(l.idx, "new.txt") && !hasName(l.idx, "old.txt")
	})
}

func TestLazyCloseWithoutOpening(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	l := NewLazy([]string{t.TempDir()}, config.Filter{Mode: "exclude"})

	if err := l.Close(); err != nil {
		t.Errorf("closing an unopened index: %v", err)
	}
	if entries := l.Entries(); entries != nil && l.idx == nil {
		t.Error("entries came from nowhere")
	}
}
