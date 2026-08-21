package i18n

import "testing"

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
