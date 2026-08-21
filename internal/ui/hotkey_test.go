package ui

import (
	"testing"

	"golang.design/x/hotkey"
)

func TestParseHotkey(t *testing.T) {
	mods, key, err := parseHotkey("ctrl+space")
	if err != nil {
		t.Fatalf("parseHotkey: %v", err)
	}
	if len(mods) != 1 || mods[0] != hotkey.ModCtrl {
		t.Errorf("mods = %v, want [ModCtrl]", mods)
	}
	if key != hotkey.KeySpace {
		t.Errorf("key = %v, want KeySpace", key)
	}

	mods, key, err = parseHotkey("Super+Shift+P")
	if err != nil {
		t.Fatalf("parseHotkey: %v", err)
	}
	if len(mods) != 2 {
		t.Errorf("mods = %v, want two modifiers", mods)
	}
	if key != hotkey.KeyP {
		t.Errorf("key = %v, want KeyP", key)
	}
}

func TestParseHotkeyRejectsInvalidSpecs(t *testing.T) {
	for _, spec := range []string{"", "space", "bogus+space", "ctrl+bogus"} {
		if _, _, err := parseHotkey(spec); err == nil {
			t.Errorf("parseHotkey(%q) succeeded, want error", spec)
		}
	}
}
