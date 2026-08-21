// Package i18n wraps Fyne's translation support behind a tiny API so
// the rest of the app never depends on the engine directly.
package i18n

import (
	"embed"

	"fyne.io/fyne/v2/lang"
)

//go:embed translations
var translations embed.FS

// Init loads the embedded translation bundles. language is "system"
// to follow the OS locale, or a code such as "en" or "es".
func Init(language string) error {
	if err := lang.AddTranslationsFS(translations, "translations"); err != nil {
		return err
	}
	if language == "" || language == "system" {
		return nil
	}
	// Fyne matches bundles against the system locale only, so forcing a
	// language means mapping its bundle onto the current locale.
	data, err := translations.ReadFile("translations/" + language + ".json")
	if err != nil {
		return err
	}
	return lang.AddTranslationsForLocale(data, lang.SystemLocale())
}

// T returns the translation for s, falling back to s itself (English).
func T(s string) string { return lang.L(s) }
