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
			mtime INTEGER NOT NULL,
			kind  INTEGER NOT NULL DEFAULT 1
		)`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			db.Close()
			return nil, err
		}
	}
	if err := migrateKindColumn(db); err != nil {
		db.Close()
		return nil, err
	}
	return &store{db: db}, nil
}

// migrateKindColumn adds the kind column to a database created before
// directories were indexed. CREATE TABLE IF NOT EXISTS is a no-op on
// an existing table, so an old three-column database needs an
// explicit ALTER; new databases already have the column and are left
// untouched. Existing rows default to KindFile (1), which is correct
// since only files were ever stored before this migration.
func migrateKindColumn(db *sql.DB) error {
	rows, err := db.Query(`PRAGMA table_info(files)`)
	if err != nil {
		return err
	}
	hasKind := false
	for rows.Next() {
		var cid, notNull, pk int
		var name, colType string
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk); err != nil {
			rows.Close()
			return err
		}
		if name == "kind" {
			hasKind = true
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	// Drain and close before altering the table: the pragma query and
	// the ALTER must not overlap on the same connection.
	if err := rows.Close(); err != nil {
		return err
	}
	if hasKind {
		return nil
	}

	_, err = db.Exec(`ALTER TABLE files ADD COLUMN kind INTEGER NOT NULL DEFAULT 1`)
	return err
}

func (s *store) loadAll() ([]launcher.Entry, error) {
	rows, err := s.db.Query(`SELECT path, name, mtime, kind FROM files`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []launcher.Entry
	for rows.Next() {
		var e launcher.Entry
		var mtime int64
		var kind int
		if err := rows.Scan(&e.Path, &e.Name, &mtime, &kind); err != nil {
			return nil, err
		}
		e.Kind = launcher.Kind(kind)
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
	stmt, err := tx.Prepare(`INSERT INTO files (path, name, mtime, kind) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, e := range entries {
		if _, err := stmt.Exec(e.Path, e.Name, e.ModTime.Unix(), int(e.Kind)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *store) upsert(e launcher.Entry) error {
	_, err := s.db.Exec(`INSERT INTO files (path, name, mtime, kind) VALUES (?, ?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET name = excluded.name, mtime = excluded.mtime, kind = excluded.kind`,
		e.Path, e.Name, e.ModTime.Unix(), int(e.Kind))
	return err
}

// deletePrefix removes a path and, when it was a directory, its whole
// subtree.
func (s *store) deletePrefix(path string) error {
	_, err := s.db.Exec(`DELETE FROM files WHERE path = ? OR path LIKE ? || '/%'`, path, path)
	return err
}

func (s *store) close() error { return s.db.Close() }
