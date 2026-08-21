//go:build windows || darwin || android

package systray

// SetIconName is a no-op on platforms whose tray hosts render the
// icon set through SetIcon; it exists so callers build everywhere.
//
// laucha patch: not part of upstream fyne.io/systray v1.12.2.
func SetIconName(name string) bool { return true }
