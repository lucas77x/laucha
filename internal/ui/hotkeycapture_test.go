package ui

import (
	"testing"

	"fyne.io/fyne/v2"
)

func TestCaptureKeyTokenMapsToParserVocabulary(t *testing.T) {
	cases := []struct {
		key  fyne.KeyName
		want string
		ok   bool
	}{
		{fyne.KeySpace, "space", true},
		{fyne.KeyA, "a", true},
		{fyne.Key7, "7", true},
		{fyne.KeyF2, "f2", true},
		{fyne.KeyReturn, "return", true},
		{fyne.KeyPageDown, "", false},
	}
	for _, c := range cases {
		got, ok := captureKeyToken(c.key)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("captureKeyToken(%q) = %q, %v; want %q, %v", c.key, got, ok, c.want, c.ok)
		}
	}
}

func TestHeldModifiersCanonicalOrder(t *testing.T) {
	h := &hotkeyCapture{mods: map[string]bool{"super": true, "ctrl": true}}

	got := h.heldModifiers()

	if len(got) != 2 || got[0] != "ctrl" || got[1] != "super" {
		t.Errorf("heldModifiers = %v, want [ctrl super]", got)
	}
}
