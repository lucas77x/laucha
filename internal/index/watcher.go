package index

import (
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
)

// watcher keeps the index in sync with filesystem events, so a file
// created seconds ago is already searchable.
type watcher struct {
	fs  *fsnotify.Watcher
	idx *Index
}

func newWatcher(idx *Index) (*watcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &watcher{fs: fw, idx: idx}, nil
}

// watchDirs registers every directory; failures (typically the
// inotify watch limit) only degrade freshness, never break search.
func (w *watcher) watchDirs(dirs []string) {
	failed := 0
	for _, dir := range dirs {
		if err := w.fs.Add(dir); err != nil {
			failed++
		}
	}
	if failed > 0 {
		log.Printf("index: %d directories not watched (consider raising fs.inotify.max_user_watches)", failed)
	}
}

func (w *watcher) run() {
	for {
		select {
		case event, ok := <-w.fs.Events:
			if !ok {
				return
			}
			w.handle(event)
		case err, ok := <-w.fs.Errors:
			if !ok {
				return
			}
			log.Printf("index: watcher: %v", err)
		}
	}
}

func (w *watcher) handle(event fsnotify.Event) {
	switch {
	case event.Op.Has(fsnotify.Create):
		info, err := os.Stat(event.Name)
		if err != nil {
			return
		}
		if info.IsDir() {
			filter := w.idx.currentFilter()
			if filter.EnterDir(event.Name) {
				files, dirs := walk([]string{event.Name}, filter)
				for _, e := range files {
					w.idx.add(e.Path)
				}
				w.watchDirs(dirs)
			}
			return
		}
		w.idx.add(event.Name)
	case event.Op.Has(fsnotify.Write):
		w.idx.add(event.Name)
	case event.Op.Has(fsnotify.Remove), event.Op.Has(fsnotify.Rename):
		w.idx.remove(event.Name)
	}
}

func (w *watcher) close() { w.fs.Close() }
