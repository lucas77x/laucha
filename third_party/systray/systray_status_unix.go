//go:build (linux || freebsd || openbsd || netbsd) && !android

package systray

import (
	"fyne.io/systray/internal/generated/notifier"
)

// SetStatus publishes the StatusNotifierItem status: "Active" shows
// the tray item, "Passive" hides it in SNI hosts. It reports whether
// the tray connection was ready.
//
// laucha patch: not part of upstream fyne.io/systray v1.12.2.
func SetStatus(status string) bool {
	instance.lock.Lock()
	props := instance.props
	conn := instance.conn
	instance.lock.Unlock()

	if props == nil {
		return false
	}
	props.SetMust("org.kde.StatusNotifierItem", "Status", status)
	if conn == nil {
		return false
	}
	_ = notifier.Emit(conn, &notifier.StatusNotifierItem_NewStatusSignal{
		Path: path,
		Body: &notifier.StatusNotifierItem_NewStatusSignalBody{Status: status},
	})
	return true
}
