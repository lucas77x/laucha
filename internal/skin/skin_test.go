package skin

import (
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

func TestParseColor(t *testing.T) {
	cases := []struct {
		in   string
		want color.NRGBA
		ok   bool
	}{
		{"#1B191F", color.NRGBA{R: 0x1B, G: 0x19, B: 0x1F, A: 0xFF}, true},
		{"#e8a0b4", color.NRGBA{R: 0xE8, G: 0xA0, B: 0xB4, A: 0xFF}, true},
		{"#11223344", color.NRGBA{R: 0x11, G: 0x22, B: 0x33, A: 0x44}, true},
		{"1B191F", color.NRGBA{}, false},
		{"#XYZ", color.NRGBA{}, false},
		{"#1B191", color.NRGBA{}, false},
	}
	for _, c := range cases {
		got, ok := ParseColor(c.in)
		if ok != c.ok || got != c.want {
			t.Errorf("ParseColor(%q) = %v, %v; want %v, %v", c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestLoadFallsBackToClassic(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	s, err := Load("classic")
	if err != nil {
		t.Fatalf("Load(classic): %v", err)
	}
	if s.Colors.Accent != Default().Colors.Accent {
		t.Errorf("Accent = %s, want built-in default", s.Colors.Accent)
	}

	if _, err := Load("missing-skin"); err == nil {
		t.Error("Load(missing-skin) succeeded, want error")
	}
}

func TestLoadReadsUserSkinWithDefaults(t *testing.T) {
	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", base)

	dir := filepath.Join(base, "laucha", "skins", "fortnite")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	toml := []byte("name = \"Fortnite\"\n[colors]\naccent = \"#00FF88\"\n")
	if err := os.WriteFile(filepath.Join(dir, "skin.toml"), toml, 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := Load("fortnite")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if s.Name != "Fortnite" || s.Colors.Accent != "#00FF88" {
		t.Errorf("loaded %q accent %s, want Fortnite #00FF88", s.Name, s.Colors.Accent)
	}
	if s.Colors.Background != Default().Colors.Background {
		t.Errorf("unset colors must keep classic defaults, got %s", s.Colors.Background)
	}
	if s.Dir != dir {
		t.Errorf("Dir = %s, want %s", s.Dir, dir)
	}

	names := Available()
	found := false
	for _, n := range names {
		if n == "fortnite" {
			found = true
		}
	}
	if !found || names[0] != "classic" {
		t.Errorf("Available() = %v, want classic first and fortnite listed", names)
	}
}
