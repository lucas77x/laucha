package ui

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2/theme"

	"github.com/lucas77x/laucha/internal/skin"
)

func TestSkinThemeMapsDeclaredColors(t *testing.T) {
	dark := newSkinTheme(skin.Default())

	if got := dark.Color(theme.ColorNameBackground, 0); got != (color.NRGBA{R: 0x1B, G: 0x19, B: 0x1F, A: 0xFF}) {
		t.Errorf("background = %v, want the declared charcoal", got)
	}
	if got := dark.Color(theme.ColorNamePrimary, 0); got != (color.NRGBA{R: 0xE8, G: 0xA0, B: 0xB4, A: 0xFF}) {
		t.Errorf("primary = %v, want the ear rose", got)
	}
}

func TestSkinThemeAutoContrastOnAccent(t *testing.T) {
	dark := newSkinTheme(skin.Default()) // light rose accent -> dark ink text
	onAccent := dark.Color(theme.ColorNameForegroundOnPrimary, 0).(color.NRGBA)
	if luminance(onAccent) > 128 {
		t.Errorf("onAccent over light rose = %v, want a dark ink", onAccent)
	}

	custom := skin.Default()
	custom.Colors.Accent = "#20242C" // dark accent -> light text
	darkAccent := newSkinTheme(custom)
	onDark := darkAccent.Color(theme.ColorNameForegroundOnPrimary, 0).(color.NRGBA)
	if luminance(onDark) < 128 {
		t.Errorf("onAccent over dark accent = %v, want a light color", onDark)
	}

	custom.Colors.OnAccent = "#123456"
	overridden := newSkinTheme(custom)
	if got := overridden.Color(theme.ColorNameForegroundOnPrimary, 0); got != (color.NRGBA{R: 0x12, G: 0x34, B: 0x56, A: 0xFF}) {
		t.Errorf("on_accent override ignored: %v", got)
	}
}

func TestSkinThemeVariantFollowsBackground(t *testing.T) {
	if v := newSkinTheme(skin.Default()).variant; v != theme.VariantDark {
		t.Errorf("dark skin variant = %v, want dark", v)
	}
	if v := newSkinTheme(skin.DefaultLight()).variant; v != theme.VariantLight {
		t.Errorf("light skin variant = %v, want light", v)
	}
}

func TestSkinThemeDerivedColorsAndSize(t *testing.T) {
	st := newSkinTheme(skin.Default())

	button := st.Color(theme.ColorNameButton, 0).(color.NRGBA)
	background := st.Color(theme.ColorNameBackground, 0).(color.NRGBA)
	if luminance(button) <= luminance(background) {
		t.Error("button on a dark skin must be lighter than the background")
	}
	if got := st.Size(theme.SizeNameText); got != 15 {
		t.Errorf("text size = %v, want the skin's 15", got)
	}
}

func TestShiftClampsChannels(t *testing.T) {
	white := color.NRGBA{R: 250, G: 250, B: 250, A: 255}
	if got := shift(white, 0.5); got.R != 255 {
		t.Errorf("shift up from near-white = %v, want clamped 255", got)
	}
	black := color.NRGBA{R: 5, G: 5, B: 5, A: 255}
	if got := shift(black, -0.5); got.R != 0 {
		t.Errorf("shift down from near-black = %v, want clamped 0", got)
	}
}
