package index

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"

	"github.com/lucas77x/laucha/internal/launcher"
)

// store persists the file index so startups search instantly while
// the background reconcile walk refreshes the data.
type store struct {
	db *sql.DB
}

func openStore(path string) (*store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	stmts := []string{
		`PRAGMA journal_mode = WAL`,
		`CREATE TABLE IF NOT EXISTS files (
			path  TEXT PRIMARY KEY,
			name  TEXT NOT NULL,
			mtime INTEGER NOT NULL
		)`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			db.Close()
			return nil, err
		}
	}
	return &store{db: db}, nil
}

func (s *store) loadAll() ([]launcher.Entry, error) {
	rows, err := s.db.Query(`SELECT path, name, mtime FROM files`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []launcher.Entry
	for rows.Next() {
		var e launcher.Entry
		var mtime int64
		if err := rows.Scan(&e.Path, &e.Name, &mtime); err != nil {
			return nil, err
		}
		e.Kind = launcher.KindFile
		e.ModTime = time.Unix(mtime, 0)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// replaceAll swaps the whole table for a fresh walk result in one
// transaction.
func (s *store) replaceAll(entries []launcher.Entry) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM files`); err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO files (path, name, mtime) VALUES (?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, e := range entries {
		if _, err := stmt.Exec(e.Path, e.Name, e.ModTime.Unix()); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *store) upsert(e launcher.Entry) error {
	_, err := s.db.Exec(`INSERT INTO files (path, name, mtime) VALUES (?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET name = excluded.name, mtime = excluded.mtime`,
		e.Path, e.Name, e.ModTime.Unix())
	return err
}

// deletePrefix removes a path and, when it was a directory, its whole
// subtree.
func (s *store) deletePrefix(path string) error {
	_, err := s.db.Exec(`DELETE FROM files WHERE path = ? OR path LIKE ? || '/%'`, path, path)
	return err
}

func (s *store) close() error { return s.db.Close() }
