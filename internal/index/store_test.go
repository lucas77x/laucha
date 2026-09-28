package index

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/lucas77x/laucha/internal/launcher"
)

func TestStoreRoundTrip(t *testing.T) {
	st, err := openStore(filepath.Join(t.TempDir(), "index.db"))
	if err != nil {
		t.Fatalf("openStore: %v", err)
	}
	defer st.close()

	entries := []launcher.Entry{
		{Kind: launcher.KindFile, Name: "a.txt", Path: "/x/a.txt", ModTime: time.Unix(100, 0)},
		{Kind: launcher.KindFile, Name: "b.pdf", Path: "/x/sub/b.pdf", ModTime: time.Unix(200, 0)},
	}
	if err := st.replaceAll(entries); err != nil {
		t.Fatalf("replaceAll: %v", err)
	}
	if err := st.deletePrefix("/x/sub"); err != nil {
		t.Fatalf("deletePrefix: %v", err)
	}

	got, err := st.loadAll()
	if err != nil {
		t.Fatalf("loadAll: %v", err)
	}
	if len(got) != 1 || got[0].Path != "/x/a.txt" {
		t.Errorf("loadAll = %+v, want only /x/a.txt", got)
	}
	if got[0].ModTime.Unix() != 100 {
		t.Errorf("ModTime = %d, want 100", got[0].ModTime.Unix())
	}
}

func TestStoreKindRoundTrip(t *testing.T) {
	st, err := openStore(filepath.Join(t.TempDir(), "index.db"))
	if err != nil {
		t.Fatalf("openStore: %v", err)
	}
	defer st.close()

	entries := []launcher.Entry{
		{Kind: launcher.KindFile, Name: "a.txt", Path: "/x/a.txt", ModTime: time.Unix(100, 0)},
		{Kind: launcher.KindDir, Name: "sub", Path: "/x/sub", ModTime: time.Unix(200, 0)},
	}
	if err := st.replaceAll(entries); err != nil {
		t.Fatalf("replaceAll: %v", err)
	}

	got, err := st.loadAll()
	if err != nil {
		t.Fatalf("loadAll: %v", err)
	}
	byPath := map[string]launcher.Entry{}
	for _, e := range got {
		byPath[e.Path] = e
	}
	if byPath["/x/a.txt"].Kind != launcher.KindFile {
		t.Errorf("a.txt kind = %v, want KindFile", byPath["/x/a.txt"].Kind)
	}
	if byPath["/x/sub"].Kind != launcher.KindDir {
		t.Errorf("sub kind = %v, want KindDir", byPath["/x/sub"].Kind)
	}

	// upsert must also round-trip kind.
	if err := st.upsert(launcher.Entry{Kind: launcher.KindDir, Name: "new", Path: "/x/new", ModTime: time.Unix(300, 0)}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	got, err = st.loadAll()
	if err != nil {
		t.Fatalf("loadAll: %v", err)
	}
	found := false
	for _, e := range got {
		if e.Path == "/x/new" {
			found = true
			if e.Kind != launcher.KindDir {
				t.Errorf("upserted kind = %v, want KindDir", e.Kind)
			}
		}
	}
	if !found {
		t.Error("upserted entry not found")
	}
}

func TestStoreMigratesOldSchema(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "index.db")

	// Simulate a database created before the kind column existed.
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	if _, err := db.Exec(`CREATE TABLE files (
		path  TEXT PRIMARY KEY,
		name  TEXT NOT NULL,
		mtime INTEGER NOT NULL
	)`); err != nil {
		t.Fatalf("create old schema: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO files (path, name, mtime) VALUES (?, ?, ?)`,
		"/x/old.txt", "old.txt", int64(50)); err != nil {
		t.Fatalf("insert old row: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	st, err := openStore(dbPath)
	if err != nil {
		t.Fatalf("openStore on old schema: %v", err)
	}
	defer st.close()

	got, err := st.loadAll()
	if err != nil {
		t.Fatalf("loadAll after migration: %v", err)
	}
	if len(got) != 1 || got[0].Path != "/x/old.txt" {
		t.Fatalf("loadAll = %+v, want only /x/old.txt", got)
	}
	if got[0].Kind != launcher.KindFile {
		t.Errorf("migrated row kind = %v, want KindFile (default)", got[0].Kind)
	}

	// The migrated database must also accept new writes with a kind.
	if err := st.upsert(launcher.Entry{Kind: launcher.KindDir, Name: "d", Path: "/x/d", ModTime: time.Unix(60, 0)}); err != nil {
		t.Fatalf("upsert after migration: %v", err)
	}
}
