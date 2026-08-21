// Package ui builds the launcher bar: a borderless window with a
// search input on top and the ranked results below.
package ui

import (
	"errors"
	"log"
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

// UsageRecorder counts opens so the ranking can favor what the user
// actually launches.
type UsageRecorder interface {
	Record(path string) error
}

// Deps groups the collaborators the bar needs; Recents and Usage are
// optional.
type Deps struct {
	Engine  *search.Engine
	Recents RecentSource
	Usage   UsageRecorder
}

type Bar struct {
	app      fyne.App
	win      fyne.Window
	cfg      config.Config
	engine   *search.Engine
	recents  RecentSource
	usage    UsageRecorder
	input    *searchEntry
	list     *widget.List
	results  []launcher.Entry
	selected int
	resident bool // tray or hotkey active: hide instead of quitting
	visible  bool
}

func New(cfg config.Config, deps Deps) *Bar {
	appIcon := fyne.NewStaticResource("icon.svg", assets.IconSVG)
	app.SetMetadata(fyne.AppMetadata{
		ID:      "com.github.lucas77x.laucha",
		Name:    "laucha",
		Version: "0.1.0",
		Icon:    appIcon,
	})
	b := &Bar{
		app:     app.NewWithID("com.github.lucas77x.laucha"),
		cfg:     cfg,
		engine:  deps.Engine,
		recents: deps.Recents,
		usage:   deps.Usage,
	}
	b.app.SetIcon(appIcon)
	b.win = b.newWindow()
	b.win.SetTitle("laucha")
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

// Run shows the bar and blocks until the app exits. With a tray icon
// or a registered hotkey the app stays resident: closing the bar
// only hides it.
func (b *Bar) Run() {
	trayOK := b.setupTray()
	hotkeyOK := b.registerHotkey()
	b.resident = trayOK || hotkeyOK

	b.win.SetCloseIntercept(func() {
		if b.resident && b.cfg.Behavior.MinimizeOnClose {
			b.hide()
			return
		}
		b.app.Quit()
	})
	b.app.Lifecycle().SetOnExitedForeground(func() {
		if b.cfg.Behavior.HideOnFocusLost && b.visible {
			b.hide()
		}
	})

	b.visible = true
	b.win.Canvas().Focus(b.input)
	b.win.ShowAndRun()
}

// show resets the query and brings the bar to the front focused.
func (b *Bar) show() {
	b.input.SetText("")
	b.search("") // refresh recents even when the text was already empty
	b.win.Show()
	b.win.RequestFocus()
	b.win.Canvas().Focus(b.input)
	b.visible = true
}

// hide keeps the app resident when possible, otherwise quits.
func (b *Bar) hide() {
	if !b.resident {
		b.app.Quit()
		return
	}
	b.visible = false
	b.win.Hide()
}

func (b *Bar) toggle() {
	if b.visible {
		b.hide()
		return
	}
	b.show()
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
			path.TextStyle = fyne.TextStyle{Italic: true}
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
		b.hide()
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
	entry := b.results[b.selected]
	if err := open(entry); err != nil {
		return
	}
	if b.usage != nil {
		if err := b.usage.Record(entry.Path); err != nil {
			log.Printf("usage: recording open: %v", err)
		}
	}
	b.hide()
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
