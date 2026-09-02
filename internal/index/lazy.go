package index

import (
	"log"
	"sync"

	"github.com/lucas77x/laucha/internal/config"
	"github.com/lucas77x/laucha/internal/launcher"
)

// Lazy opens the index the first time file results are asked for.
//
// The bar only reaches the file provider when file search is enabled,
// so a launcher configured for applications alone never opens the
// database, never walks the roots and never pays for the entries it
// would hold.
type Lazy struct {
	mu     sync.Mutex
	roots  []string
	filter config.Filter
	idx    *Index
	failed bool
}

// NewLazy records the settings the index will be opened with.
func NewLazy(roots []string, filter config.Filter) *Lazy {
	return &Lazy{roots: roots, filter: filter}
}

// index opens the real index once; a failure is logged and remembered
// so a broken database is not retried on every keystroke.
func (l *Lazy) index() *Index {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.idx != nil || l.failed {
		return l.idx
	}
	idx, err := New(l.roots, l.filter)
	if err != nil {
		log.Printf("file index disabled: %v", err)
		l.failed = true
		return nil
	}
	l.idx = idx
	return idx
}

// Entries implements search.Provider.
func (l *Lazy) Entries() []launcher.Entry {
	if idx := l.index(); idx != nil {
		return idx.Entries()
	}
	return nil
}

// Recent implements ui.RecentSource.
func (l *Lazy) Recent(n int) []launcher.Entry {
	if idx := l.index(); idx != nil {
		return idx.Recent(n)
	}
	return nil
}

// Reconfigure implements ui.Reconfigurer. Settings changed before the
// index is opened are simply remembered for when it is.
func (l *Lazy) Reconfigure(roots []string, filter config.Filter) {
	l.mu.Lock()
	l.roots = roots
	l.filter = filter
	idx := l.idx
	l.mu.Unlock()

	if idx != nil {
		idx.Reconfigure(roots, filter)
	}
}

// Opened reports whether the index was ever needed.
func (l *Lazy) Opened() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.idx != nil
}

func (l *Lazy) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.idx == nil {
		return nil
	}
	return l.idx.Close()
}
