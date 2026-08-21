// Package index maintains a live index of the user's files: loaded
// from SQLite at startup, reconciled by a background walk and kept
// fresh by filesystem watchers. It implements search.Provider.
//
// Watcher goroutines only mutate the index, never the UI, so no
// fyne.Do bridging is needed here.
package index

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/lucas77x/laucha/internal/config"
	"github.com/lucas77x/laucha/internal/launcher"
)

type Index struct {
	filter *Filter
	roots  []string
	store  *store

	mu       sync.RWMutex
	byPath   map[string]launcher.Entry
	snapshot []launcher.Entry // rebuilt lazily; readers must not mutate
	dirty    bool

	watcher *watcher
}

// New opens the on-disk index and starts the background reconcile
// walk and the filesystem watchers. Search works immediately on the
// previously indexed data.
func New(search config.Search, filterCfg config.Filter) (*Index, error) {
	dbPath, err := dataPath()
	if err != nil {
		return nil, err
	}
	st, err := openStore(dbPath)
	if err != nil {
		return nil, err
	}

	idx := &Index{
		filter: NewFilter(filterCfg),
		roots:  expandRoots(search.Roots),
		store:  st,
		byPath: map[string]launcher.Entry{},
		dirty:  true,
	}
	entries, err := st.loadAll()
	if err != nil {
		log.Printf("index: loading stored index: %v", err)
	}
	for _, e := range entries {
		idx.byPath[e.Path] = e
	}

	go idx.reconcile()
	return idx, nil
}

// Entries implements search.Provider. The returned slice is shared:
// callers must treat it as read-only.
func (i *Index) Entries() []launcher.Entry {
	i.mu.RLock()
	if !i.dirty {
		snapshot := i.snapshot
		i.mu.RUnlock()
		return snapshot
	}
	i.mu.RUnlock()

	i.mu.Lock()
	defer i.mu.Unlock()
	if i.dirty {
		i.snapshot = make([]launcher.Entry, 0, len(i.byPath))
		for _, e := range i.byPath {
			i.snapshot = append(i.snapshot, e)
		}
		i.dirty = false
	}
	return i.snapshot
}

// Recent returns up to n files, newest modification first.
func (i *Index) Recent(n int) []launcher.Entry {
	entries := i.Entries()
	recent := make([]launcher.Entry, len(entries))
	copy(recent, entries)
	sort.Slice(recent, func(a, b int) bool {
		return recent[a].ModTime.After(recent[b].ModTime)
	})
	if len(recent) > n {
		recent = recent[:n]
	}
	return recent
}

func (i *Index) Close() error {
	if i.watcher != nil {
		i.watcher.close()
	}
	return i.store.close()
}

// reconcile replaces memory and store with a fresh walk while the UI
// keeps searching the previously loaded data.
func (i *Index) reconcile() {
	files, dirs := walk(i.roots, i.filter)

	i.mu.Lock()
	i.byPath = make(map[string]launcher.Entry, len(files))
	for _, e := range files {
		i.byPath[e.Path] = e
	}
	i.dirty = true
	i.mu.Unlock()

	if err := i.store.replaceAll(files); err != nil {
		log.Printf("index: persisting: %v", err)
	}

	w, err := newWatcher(i)
	if err != nil {
		log.Printf("index: watcher disabled: %v", err)
		return
	}
	i.watcher = w
	w.watchDirs(dirs)
	go w.run()
}

func (i *Index) add(path string) {
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return
	}
	if !i.filter.IncludeFile(path) {
		return
	}
	entry := launcher.Entry{
		Kind:    launcher.KindFile,
		Name:    filepath.Base(path),
		Path:    path,
		ModTime: info.ModTime(),
	}

	i.mu.Lock()
	i.byPath[path] = entry
	i.dirty = true
	i.mu.Unlock()

	if err := i.store.upsert(entry); err != nil {
		log.Printf("index: upsert %s: %v", path, err)
	}
}

func (i *Index) remove(path string) {
	prefix := path + "/"

	i.mu.Lock()
	for p := range i.byPath {
		if p == path || strings.HasPrefix(p, prefix) {
			delete(i.byPath, p)
		}
	}
	i.dirty = true
	i.mu.Unlock()

	if err := i.store.deletePrefix(path); err != nil {
		log.Printf("index: delete %s: %v", path, err)
	}
}

func dataPath() (string, error) {
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".local", "share")
	}
	dir := filepath.Join(base, "laucha")
	// 0700: the index lists every filename under the user's roots and
	// must not be readable by other local users.
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "index.db"), nil
}
