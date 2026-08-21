// Package ui builds the launcher bar: a borderless window with a
// search input on top and the ranked results below.
package ui

import (
	"errors"
	"os/exec"
	"strings"
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"github.com/lucas77x/laucha/assets"
	"github.com/lucas77x/laucha/internal/config"
	"github.com/lucas77x/laucha/internal/i18n"
	"github.com/lucas77x/laucha/internal/launcher"
	"github.com/lucas77x/laucha/internal/search"
)

const (
	rowHeight  = 44 // refined by the skin engine later
	iconSize   = 28
	queryLimit = 50 // results kept for scrolling beyond the visible rows
)

// RecentSource lists recently modified files for the empty-query
// view.
type RecentSource interface {
	Recent(n int) []launcher.Entry
}

type Bar struct {
	app      fyne.App
	win      fyne.Window
	cfg      config.Config
	engine   *search.Engine
	recents  RecentSource
	input    *searchEntry
	list     *widget.List
	results  []launcher.Entry
	selected int
}

func New(cfg config.Config, engine *search.Engine, recents RecentSource) *Bar {
	b := &Bar{
		app:     app.NewWithID("com.github.lucas77x.laucha"),
		cfg:     cfg,
		engine:  engine,
		recents: recents,
	}
	b.app.SetIcon(fyne.NewStaticResource("icon.svg", assets.IconSVG))
	b.win = b.newWindow()
	b.input = newSearchEntry(b.handleKey)
	b.input.SetPlaceHolder(i18n.T("Search apps and files…"))
	b.input.OnChanged = b.search
	b.list = b.newList()
	b.list.OnSelected = func(id widget.ListItemID) { b.selected = id }

	b.win.SetContent(container.NewBorder(b.input, nil, nil, nil, b.list))
	height := b.input.MinSize().Height + float32(cfg.Window.MaxItems)*rowHeight
	b.win.Resize(fyne.NewSize(cfg.Window.Width, height))
	b.win.CenterOnScreen()
	b.search("")
	return b
}

// Run shows the bar and blocks until the app exits.
func (b *Bar) Run() {
	b.win.Canvas().Focus(b.input)
	b.win.ShowAndRun()
}

// newWindow prefers a splash window: borderless and centered, the
// natural shape for a launcher bar.
func (b *Bar) newWindow() fyne.Window {
	if drv, ok := b.app.Driver().(desktop.Driver); ok {
		return drv.CreateSplashWindow()
	}
	return b.app.NewWindow("laucha")
}

func (b *Bar) newList() *widget.List {
	return widget.NewList(
		func() int { return len(b.results) },
		func() fyne.CanvasObject {
			icon := canvas.NewImageFromFile("")
			icon.FillMode = canvas.ImageFillContain
			icon.SetMinSize(fyne.NewSize(iconSize, iconSize))
			path := widget.NewLabel("")
			path.Importance = widget.LowImportance
			return container.NewHBox(icon, widget.NewLabel(""), path)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id >= len(b.results) {
				return
			}
			entry := b.results[id]
			row := item.(*fyne.Container)
			icon := row.Objects[0].(*canvas.Image)
			name := row.Objects[1].(*widget.Label)
			path := row.Objects[2].(*widget.Label)
			if entry.Kind == launcher.KindFile {
				icon.File = ""
				icon.Resource = fileIcon(entry.Name)
				path.SetText(displayDir(entry.Path))
			} else {
				icon.Resource = nil
				icon.File = entry.Icon
				path.SetText("")
			}
			icon.Refresh()
			name.SetText(entry.Name)
		},
	)
}

func (b *Bar) search(query string) {
	if strings.TrimSpace(query) == "" {
		b.results = b.recentFiles()
	} else {
		b.results = b.engine.Query(query, queryLimit)
	}
	b.selected = 0
	b.list.Refresh()
	if len(b.results) > 0 {
		b.list.Select(0)
	} else {
		b.list.UnselectAll()
	}
}

// recentFiles feeds the empty-query view when the config enables it.
func (b *Bar) recentFiles() []launcher.Entry {
	if b.recents == nil || !b.cfg.Behavior.ShowRecentOnOpen {
		return nil
	}
	return b.recents.Recent(queryLimit)
}

// handleKey intercepts navigation keys before the entry consumes
// them; the return value reports whether the key was handled.
func (b *Bar) handleKey(key *fyne.KeyEvent) bool {
	switch key.Name {
	case fyne.KeyEscape:
		// Hiding instead of quitting arrives together with the tray
		// icon and the global hotkey.
		b.app.Quit()
		return true
	case fyne.KeyDown:
		b.moveSelection(1)
		return true
	case fyne.KeyUp:
		b.moveSelection(-1)
		return true
	case fyne.KeyReturn, fyne.KeyEnter:
		b.openSelected()
		return true
	}
	return false
}

func (b *Bar) moveSelection(delta int) {
	if len(b.results) == 0 {
		return
	}
	b.selected = (b.selected + delta + len(b.results)) % len(b.results)
	b.list.Select(b.selected)
}

func (b *Bar) openSelected() {
	if b.selected >= len(b.results) {
		return
	}
	if err := open(b.results[b.selected]); err != nil {
		return
	}
	b.app.Quit()
}

// open launches an app detached from laucha, or a file with the
// desktop's default handler. Commands run directly, never through a
// shell.
func open(entry launcher.Entry) error {
	var cmd *exec.Cmd
	switch entry.Kind {
	case launcher.KindApp:
		if len(entry.Exec) == 0 {
			return errors.New("entry has no command")
		}
		cmd = exec.Command(entry.Exec[0], entry.Exec[1:]...)
	default:
		cmd = exec.Command("xdg-open", entry.Path)
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	return cmd.Start()
}

// searchEntry lets the bar intercept navigation keys while normal
// typing still goes to the underlying entry.
type searchEntry struct {
	widget.Entry
	onKey func(*fyne.KeyEvent) bool
}

func newSearchEntry(onKey func(*fyne.KeyEvent) bool) *searchEntry {
	e := &searchEntry{onKey: onKey}
	e.ExtendBaseWidget(e)
	return e
}

func (e *searchEntry) TypedKey(key *fyne.KeyEvent) {
	if e.onKey(key) {
		return
	}
	e.Entry.TypedKey(key)
}
