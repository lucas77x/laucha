package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// forcedVariant pins the light or dark variant regardless of the
// system setting.
type forcedVariant struct {
	fyne.Theme
	variant fyne.ThemeVariant
}

func (f *forcedVariant) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return f.Theme.Color(name, f.variant)
}

// applyTheme applies the loaded skin immediately; when the skin
// cannot be loaded it falls back to the plain light/dark setting.
func (b *Bar) applyTheme() {
	if b.skin.Name != "" {
		b.app.Settings().SetTheme(newSkinTheme(b.skin))
		return
	}
	switch b.cfg.Window.Theme {
	case "light":
		b.app.Settings().SetTheme(&forcedVariant{Theme: theme.DefaultTheme(), variant: theme.VariantLight})
	case "dark":
		b.app.Settings().SetTheme(&forcedVariant{Theme: theme.DefaultTheme(), variant: theme.VariantDark})
	default:
		b.app.Settings().SetTheme(theme.DefaultTheme())
	}
}
