package index

import (
	"path/filepath"
	"testing"
	"time"

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
