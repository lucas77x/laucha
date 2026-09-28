package i18n

import (
	"os"
	"testing"

	"fyne.io/fyne/v2/lang"
)

// Fyne's lang package keeps its bundle and active localizer as
// package-level state, so every test below calls Init itself rather
// than relying on a previous test having set it up: order must not
// matter.

func TestInitAndTranslate(t *testing.T) {
	if err := Init("system"); err != nil {
		t.Fatalf("Init(system): %v", err)
	}
	if got := T("A key nobody translated"); got != "A key nobody translated" {
		t.Errorf("unknown key = %q, want the source text back", got)
	}
}

func TestInitForcesLanguage(t *testing.T) {
	if err := Init("es"); err != nil {
		t.Fatalf("Init(es): %v", err)
	}
	if got := T("Show"); got != "Mostrar" {
		t.Errorf("T(Show) with forced es = %q, want Mostrar", got)
	}
}

func TestInitRejectsUnknownLanguage(t *testing.T) {
	if err := Init("xx"); err == nil {
		t.Error("Init(xx) succeeded, want error for a missing bundle")
	}
}

// TestInitForcedLanguageMergesFyneStrings covers the actual bug: a
// forced language must also resolve Fyne's own bundled strings (never
// shipped by laucha), not just laucha's.
func TestInitForcedLanguageMergesFyneStrings(t *testing.T) {
	if err := Init("es"); err != nil {
		t.Fatalf("Init(es): %v", err)
	}
	if got := lang.L("Quit"); got != "Salir" {
		t.Errorf("lang.L(Quit) with forced es = %q, want Salir (Fyne's own string)", got)
	}
	if got := T("Show"); got != "Mostrar" {
		t.Errorf("T(Show) with forced es = %q, want Mostrar (laucha's own string)", got)
	}
}

// TestInitForcedLanguageOverridesEnv proves the forced language wins
// over whatever the process environment says, not just over the
// default system locale.
func TestInitForcedLanguageOverridesEnv(t *testing.T) {
	t.Setenv("LANGUAGE", "es_AR")
	if err := Init("en"); err != nil {
		t.Fatalf("Init(en): %v", err)
	}
	if got := lang.L("Quit"); got != "Quit" {
		t.Errorf("lang.L(Quit) with forced en over LANGUAGE=es_AR = %q, want Quit", got)
	}
}

// TestInitForcedLanguageOverridesBlockingLCAll covers the go-locale
// quirk where LC_ALL=C (or POSIX) short-circuits LANGUAGE entirely:
// Init must clear it long enough for the forced code to take effect.
func TestInitForcedLanguageOverridesBlockingLCAll(t *testing.T) {
	t.Setenv("LC_ALL", "C")
	if err := Init("es"); err != nil {
		t.Fatalf("Init(es): %v", err)
	}
	if got := lang.L("Quit"); got != "Salir" {
		t.Errorf("lang.L(Quit) with forced es over LC_ALL=C = %q, want Salir", got)
	}
	if got := os.Getenv("LC_ALL"); got != "C" {
		t.Errorf("LC_ALL after Init = %q, want C (restored)", got)
	}
}

// TestInitRestoresLanguageEnv checks both restore cases: a
// previously-set LANGUAGE comes back unchanged, and a previously
// unset one stays unset — so processes laucha launches afterward
// (opened files, apps) keep the user's real locale.
func TestInitRestoresLanguageEnv(t *testing.T) {
	t.Setenv("LANGUAGE", "fr")
	if err := Init("es"); err != nil {
		t.Fatalf("Init(es): %v", err)
	}
	if got := os.Getenv("LANGUAGE"); got != "fr" {
		t.Errorf("LANGUAGE after Init = %q, want fr (restored)", got)
	}

	os.Unsetenv("LANGUAGE")
	t.Cleanup(func() { os.Unsetenv("LANGUAGE") })
	if err := Init("es"); err != nil {
		t.Fatalf("Init(es): %v", err)
	}
	if v, ok := os.LookupEnv("LANGUAGE"); ok {
		t.Errorf("LANGUAGE after Init = %q, want still unset", v)
	}
}
