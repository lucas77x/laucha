// Package launcher holds the core domain types shared by providers,
// the search engine and the UI.
package launcher

import "time"

type Kind int

const (
	KindApp Kind = iota
	KindFile
)

// Entry is anything the bar can list and open.
type Entry struct {
	Kind     Kind
	Name     string    // display name
	Path     string    // file path; for apps, the .desktop file path
	Exec     []string  // launch argv, apps only; never run through a shell
	Terminal bool      // apps only: must run inside a terminal emulator
	Icon     string    // resolved icon image path, may be empty
	ModTime  time.Time // files only; drives the recent view
}
