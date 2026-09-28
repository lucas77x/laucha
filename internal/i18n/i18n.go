// Package i18n wraps Fyne's translation support behind a tiny API so
// the rest of the app never depends on the engine directly.
package i18n

import (
	"embed"
	"os"

	"fyne.io/fyne/v2/lang"
)

//go:embed translations
var translations embed.FS

// Init loads the embedded translation bundles, merging laucha's own
// strings into Fyne's for every locale they share. language is
// "system" (or empty) to follow the OS locale as usual, or a code
// such as "en" or "es" to force one regardless of it.
func Init(language string) error {
	if language == "" || language == "system" {
		return lang.AddTranslationsFS(translations, "translations")
	}
	if _, err := translations.ReadFile("translations/" + language + ".json"); err != nil {
		return err
	}

	// Fyne re-resolves the OS locale (via github.com/jeandeaual/
	// go-locale) every time translations are added, and both laucha's
	// and Fyne's own bundles are keyed by plain language tags ("es"),
	// not the possibly-region-qualified system locale ("es-AR"). So
	// forcing the resolved locale to the requested code, instead of
	// registering laucha's strings under a separate region tag,
	// lets both bundles merge under the same key.
	restore := forceLocale(language)
	defer restore()
	return lang.AddTranslationsFS(translations, "translations")
}

// T returns the translation for s, falling back to s itself (English).
func T(s string) string { return lang.L(s) }

// localeEnvVars is the priority order go-locale's getLangFromEnv
// checks: the first one that is set decides the locale, and LANGUAGE
// is only consulted afterward, provided that value isn't "C" or
// "POSIX".
var localeEnvVars = []string{"LC_ALL", "LC_MESSAGES", "LANG"}

// forceLocale makes Fyne resolve the OS locale to code for as long as
// the returned function has not been called, then exactly restores
// every environment variable it touched — unset if it was unset,
// otherwise its previous value. go-locale prefers LANGUAGE over
// LC_ALL/LC_MESSAGES/LANG, but only when none of those three is
// literally "C" or "POSIX": that value short-circuits the lookup and
// LANGUAGE is never read. Any such blocking variable is cleared first
// so LANGUAGE can take effect.
func forceLocale(code string) func() {
	var restores []func()

	for _, name := range localeEnvVars {
		value := os.Getenv(name)
		if value == "" {
			continue
		}
		if value != "C" && value != "POSIX" {
			break // not blocking: go-locale stops here too, so nothing further to clear
		}
		blocked := value
		restores = append(restores, func() { _ = os.Setenv(name, blocked) })
		_ = os.Unsetenv(name)
	}

	prev, hadLanguage := os.LookupEnv("LANGUAGE")
	restores = append(restores, func() {
		if hadLanguage {
			_ = os.Setenv("LANGUAGE", prev)
		} else {
			_ = os.Unsetenv("LANGUAGE")
		}
	})
	_ = os.Setenv("LANGUAGE", code)

	return func() {
		for i := len(restores) - 1; i >= 0; i-- {
			restores[i]()
		}
	}
}
