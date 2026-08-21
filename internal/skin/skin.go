// Package skin loads drop-in appearance packs: a folder holding a
// skin.toml that dresses one of the predefined layout templates.
package skin

import (
	"errors"
	"image/color"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type Skin struct {
	Name     string `toml:"name"`
	Author   string `toml:"author"`
	Version  string `toml:"version"`
	Template string `toml:"template"`
	Colors   Colors `toml:"colors"`
	Font     Font   `toml:"font"`
	Rows     Rows   `toml:"rows"`
	Images   Images `toml:"images"`

	Dir string `toml:"-"` // folder the skin was loaded from
}

type Colors struct {
	Background      string `toml:"background"`
	Foreground      string `toml:"foreground"`
	Muted           string `toml:"muted"`
	Accent          string `toml:"accent"`
	Selection       string `toml:"selection"`
	InputBackground string `toml:"input_background"`
}

type Font struct {
	Size float32 `toml:"size"`
}

type Rows struct {
	Height   float32 `toml:"height"`
	IconSize float32 `toml:"icon_size"`
}

type Images struct {
	Background string `toml:"background"`
}

// Default is the built-in classic look, used when no skin folder
// exists on disk: warm charcoal surfaces, warm off-white text and the
// mouse-ear rose as the single accent.
func Default() Skin {
	return Skin{
		Name:     "Classic",
		Template: "classic",
		Colors: Colors{
			Background:      "#1B191F",
			Foreground:      "#E8E4DE",
			Muted:           "#8A879B",
			Accent:          "#E8A0B4",
			Selection:       "#3A2E36",
			InputBackground: "#141317",
		},
		Font: Font{Size: 15},
		Rows: Rows{Height: 46, IconSize: 30},
	}
}

// Dirs returns where skins are searched: next to the binary first,
// then the user config directory.
func Dirs() []string {
	var dirs []string
	if exe, err := os.Executable(); err == nil {
		dirs = append(dirs, filepath.Join(filepath.Dir(exe), "skins"))
	}
	if base, err := os.UserConfigDir(); err == nil {
		dirs = append(dirs, filepath.Join(base, "laucha", "skins"))
	}
	return dirs
}

// Available lists skin names found on disk; classic is always
// present as the built-in reference skin.
func Available() []string {
	names := []string{"classic"}
	seen := map[string]bool{"classic": true}
	for _, dir := range Dirs() {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() && !seen[entry.Name()] {
				seen[entry.Name()] = true
				names = append(names, entry.Name())
			}
		}
	}
	return names
}

// Load reads <dir>/<name>/skin.toml. Unset fields keep the classic
// values, so minimal skins stay valid.
func Load(name string) (Skin, error) {
	for _, dir := range Dirs() {
		path := filepath.Join(dir, name, "skin.toml")
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		s := Default()
		if err := toml.Unmarshal(data, &s); err != nil {
			return Default(), err
		}
		s.Dir = filepath.Join(dir, name)
		return s, nil
	}
	if name == "classic" {
		return Default(), nil
	}
	return Default(), errors.New("skin not found: " + name)
}

// BackgroundImagePath resolves the optional background image inside
// the skin folder; empty when the skin declares none.
func (s Skin) BackgroundImagePath() string {
	if s.Images.Background == "" || s.Dir == "" {
		return ""
	}
	path := filepath.Join(s.Dir, s.Images.Background)
	if _, err := os.Stat(path); err != nil {
		return ""
	}
	return path
}

// ParseColor turns #RRGGBB or #RRGGBBAA into a color; ok is false on
// malformed values.
func ParseColor(hex string) (color.NRGBA, bool) {
	if len(hex) == 0 || hex[0] != '#' {
		return color.NRGBA{}, false
	}
	hex = hex[1:]
	if len(hex) != 6 && len(hex) != 8 {
		return color.NRGBA{}, false
	}
	component := func(i int) (uint8, bool) {
		hi, ok1 := nibble(hex[i])
		lo, ok2 := nibble(hex[i+1])
		return hi<<4 | lo, ok1 && ok2
	}
	r, okR := component(0)
	g, okG := component(2)
	b, okB := component(4)
	a := uint8(0xFF)
	okA := true
	if len(hex) == 8 {
		a, okA = component(6)
	}
	if !okR || !okG || !okB || !okA {
		return color.NRGBA{}, false
	}
	return color.NRGBA{R: r, G: g, B: b, A: a}, true
}

func nibble(c byte) (uint8, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	}
	return 0, false
}
