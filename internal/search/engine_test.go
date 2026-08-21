package search

import (
	"testing"

	"github.com/lucas77x/laucha/internal/launcher"
)

type fakeProvider struct {
	entries []launcher.Entry
}

func (f fakeProvider) Entries() []launcher.Entry { return f.entries }

func provider(names ...string) fakeProvider {
	entries := make([]launcher.Entry, len(names))
	for i, n := range names {
		entries[i] = launcher.Entry{Kind: launcher.KindApp, Name: n}
	}
	return fakeProvider{entries: entries}
}

func names(entries []launcher.Entry) []string {
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.Name
	}
	return out
}

func TestQueryRanksPrefixFirst(t *testing.T) {
	engine := NewEngine(provider("LibreOffice Calc", "Calculator", "Spotify"))

	got := names(engine.Query("calc", 10))

	want := []string{"Calculator", "LibreOffice Calc"}
	if len(got) != len(want) {
		t.Fatalf("Query(calc) = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Query(calc)[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestQueryIsCaseInsensitive(t *testing.T) {
	engine := NewEngine(provider("Spotify"))

	for _, q := range []string{"spo", "Spo", "SPO"} {
		got := engine.Query(q, 10)
		if len(got) != 1 || got[0].Name != "Spotify" {
			t.Errorf("Query(%q) = %v, want [Spotify]", q, names(got))
		}
	}
}

func TestQueryMatchesSubsequence(t *testing.T) {
	engine := NewEngine(provider("Firefox"))

	if got := engine.Query("ffx", 10); len(got) != 1 {
		t.Errorf("Query(ffx) = %v, want [Firefox]", names(got))
	}
}

func TestQueryEmptyReturnsNothing(t *testing.T) {
	engine := NewEngine(provider("Spotify"))

	if got := engine.Query("   ", 10); got != nil {
		t.Errorf("Query(blank) = %v, want nil", names(got))
	}
}

func TestQueryRespectsLimit(t *testing.T) {
	engine := NewEngine(provider("app one", "app two", "app three"))

	if got := engine.Query("app", 2); len(got) != 2 {
		t.Errorf("len(Query(app, 2)) = %d, want 2", len(got))
	}
}
