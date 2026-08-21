package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	"github.com/lucas77x/laucha/internal/skin"
)

// skinTheme adapts a skin's palette to Fyne's theme system. Surface
// elevation and separators are derived from the declared colors, so
// skins stay small: they name a world, laucha builds the layers.
type skinTheme struct {
	base fyne.Theme
	s    skin.Skin

	background color.NRGBA
	foreground color.NRGBA
	muted      color.NRGBA
	accent     color.NRGBA
	selection  color.NRGBA
	input      color.NRGBA
}

func newSkinTheme(s skin.Skin) *skinTheme {
	t := &skinTheme{base: theme.DefaultTheme(), s: s}
	d := skin.Default().Colors
	t.background = parseOr(s.Colors.Background, d.Background)
	t.foreground = parseOr(s.Colors.Foreground, d.Foreground)
	t.muted = parseOr(s.Colors.Muted, d.Muted)
	t.accent = parseOr(s.Colors.Accent, d.Accent)
	t.selection = parseOr(s.Colors.Selection, d.Selection)
	t.input = parseOr(s.Colors.InputBackground, s.Colors.Background)
	return t
}

func parseOr(hex, fallback string) color.NRGBA {
	if c, ok := skin.ParseColor(hex); ok {
		return c
	}
	c, _ := skin.ParseColor(fallback)
	return c
}

func (t *skinTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return t.background
	case theme.ColorNameForeground:
		return t.foreground
	case theme.ColorNameDisabled, theme.ColorNamePlaceHolder:
		return t.muted
	case theme.ColorNamePrimary, theme.ColorNameFocus, theme.ColorNameHyperlink:
		return t.accent
	case theme.ColorNameSelection:
		return t.selection
	case theme.ColorNameInputBackground:
		return t.input
	case theme.ColorNameMenuBackground, theme.ColorNameOverlayBackground:
		return shift(t.background, 0.05) // one whisper-quiet step above base
	case theme.ColorNameHover:
		return alpha(t.foreground, 0x14)
	case theme.ColorNameSeparator:
		return alpha(t.foreground, 0x1E)
	case theme.ColorNameScrollBar:
		return alpha(t.foreground, 0x40)
	}
	return t.base.Color(name, variant)
}

func (t *skinTheme) Font(style fyne.TextStyle) fyne.Resource    { return t.base.Font(style) }
func (t *skinTheme) Icon(name fyne.ThemeIconName) fyne.Resource { return t.base.Icon(name) }

func (t *skinTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNameText && t.s.Font.Size > 0 {
		return t.s.Font.Size
	}
	return t.base.Size(name)
}

// shift lightens (positive fraction) or darkens (negative) a color.
func shift(c color.NRGBA, f float64) color.NRGBA {
	adj := func(v uint8) uint8 {
		nv := float64(v) + f*255
		if nv < 0 {
			nv = 0
		}
		if nv > 255 {
			nv = 255
		}
		return uint8(nv)
	}
	return color.NRGBA{R: adj(c.R), G: adj(c.G), B: adj(c.B), A: c.A}
}

func alpha(c color.NRGBA, a uint8) color.NRGBA {
	return color.NRGBA{R: c.R, G: c.G, B: c.B, A: a}
}
