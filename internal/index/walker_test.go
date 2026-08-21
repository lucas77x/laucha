package index

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lucas77x/laucha/internal/config"
)

func TestWalkRespectsFilter(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "keep.txt"))
	mustWrite(t, filepath.Join(root, "skip.log"))
	mustWrite(t, filepath.Join(root, ".hidden", "inside.txt"))
	mustWrite(t, filepath.Join(root, "sub", "nested.txt"))

	f := NewFilter(config.Filter{
		Mode:       "exclude",
		Extensions: []string{".log"},
		Patterns:   []string{`(^|/)\.[^/]+`},
	})

	files, dirs := walk([]string{root}, f)

	got := map[string]bool{}
	for _, e := range files {
		got[e.Name] = true
	}
	if !got["keep.txt"] || !got["nested.txt"] {
		t.Errorf("expected keep.txt and nested.txt, got %v", got)
	}
	if got["skip.log"] || got["inside.txt"] {
		t.Errorf("filtered files leaked into the walk: %v", got)
	}
	if len(dirs) != 2 { // root and sub; .hidden must be pruned
		t.Errorf("dirs = %v, want root and sub only", dirs)
	}
}

func mustWrite(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}
