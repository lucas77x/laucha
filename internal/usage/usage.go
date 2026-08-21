// Package usage records how often entries are opened. It powers the
// frecency boost: better matches always win, equally good matches
// are ordered by use.
package usage

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// Store persists open counts and keeps them in memory for fast
// ranking lookups. All methods must be called from one goroutine
// (the UI event loop).
type Store struct {
	db    *sql.DB
	stats map[string]stat
}

type stat struct {
	count int
	last  time.Time
}

func Open() (*Store, error) {
	path, err := dbPath()
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS opens (
		path        TEXT PRIMARY KEY,
		count       INTEGER NOT NULL,
		last_opened INTEGER NOT NULL
	)`); err != nil {
		db.Close()
		return nil, err
	}

	s := &Store{db: db, stats: map[string]stat{}}
	rows, err := db.Query(`SELECT path, count, last_opened FROM opens`)
	if err != nil {
		db.Close()
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p string
		var count int
		var last int64
		if err := rows.Scan(&p, &count, &last); err != nil {
			db.Close()
			return nil, err
		}
		s.stats[p] = stat{count: count, last: time.Unix(last, 0)}
	}
	if err := rows.Err(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

// Record counts one open of the entry identified by path.
func (s *Store) Record(path string) error {
	st := s.stats[path]
	st.count++
	st.last = time.Now()
	s.stats[path] = st

	_, err := s.db.Exec(`INSERT INTO opens (path, count, last_opened) VALUES (?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET count = excluded.count, last_opened = excluded.last_opened`,
		path, st.count, st.last.Unix())
	return err
}

// Stats implements search.Usage.
func (s *Store) Stats(path string) (count int, lastOpened time.Time) {
	st := s.stats[path]
	return st.count, st.last
}

func (s *Store) Close() error { return s.db.Close() }

func dbPath() (string, error) {
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".local", "share")
	}
	dir := filepath.Join(base, "laucha")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "usage.db"), nil
}
