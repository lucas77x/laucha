package ui

// applyTheme applies the loaded skin's palette to every window
// immediately. Skins own the appearance: there is no separate
// light/dark setting, just default-dark and default-light skins.
func (b *Bar) applyTheme() {
	b.app.Settings().SetTheme(newSkinTheme(b.skin))
}
