//go:build (linux || freebsd || openbsd || netbsd) && !android

package systray

import (
	"log"

	"fyne.io/systray/internal/generated/notifier"
)

// SetIconName publishes the tray icon as a themed-icon name or an
// absolute file path. Tray hosts based on ayatana indicators (MATE,
// XFCE) ignore IconPixmap and only render IconName, so pixmap-only
// icons never show there. It reports whether the tray connection was
// ready; callers may retry until it is.
//
// laucha patch: not part of upstream fyne.io/systray v1.12.2.
func SetIconName(name string) bool {
	instance.lock.Lock()
	props := instance.props
	conn := instance.conn
	instance.lock.Unlock()

	if props == nil {
		return false
	}
	props.SetMust("org.kde.StatusNotifierItem", "IconName", name)
	if conn == nil {
		return false
	}
	if err := notifier.Emit(conn, &notifier.StatusNotifierItem_NewIconSignal{
		Path: path,
		Body: &notifier.StatusNotifierItem_NewIconSignalBody{},
	}); err != nil {
		log.Printf("systray error: failed to emit new icon signal: %s\n", err)
	}
	return true
}
