package index

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lucas77x/laucha/internal/config"
	"github.com/lucas77x/laucha/internal/launcher"
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

func TestWalkEmitsDirs(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "sub", "nested.txt"))
	mustWrite(t, filepath.Join(root, "excluded", "inside.txt"))

	f := NewFilter(config.Filter{
		Mode:  "exclude",
		Names: []string{"excluded"},
	})

	entries, _ := walk([]string{root}, f)

	dirNames := map[string]bool{}
	for _, e := range entries {
		if e.Kind == launcher.KindDir {
			dirNames[e.Name] = true
		}
		if e.Path == root {
			t.Error("root must not be indexed")
		}
	}
	if !dirNames["sub"] {
		t.Errorf("expected sub to be indexed as a dir, got %v", dirNames)
	}
	if dirNames["excluded"] {
		t.Errorf("excluded dir must not be indexed, got %v", dirNames)
	}
}

func TestWalkIncludeOnlyDirMatchesByName(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "docs", "readme.txt"))
	mustWrite(t, filepath.Join(root, "other", "readme.txt"))

	f := NewFilter(config.Filter{
		Mode:  "include-only",
		Names: []string{"docs"},
	})

	entries, _ := walk([]string{root}, f)

	dirNames := map[string]bool{}
	for _, e := range entries {
		if e.Kind == launcher.KindDir {
			dirNames[e.Name] = true
		}
	}
	if !dirNames["docs"] {
		t.Errorf("expected docs to be indexed as a dir, got %v", dirNames)
	}
	if dirNames["other"] {
		t.Errorf("unmatched dir must not be indexed, got %v", dirNames)
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
