// Package ui builds the launcher bar: a borderless window with a
// search input on top and the ranked results below.
package ui

import (
	"errors"
	"image/color"
	"log"
	"os/exec"
	"strings"
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/lucas77x/laucha/assets"
	"github.com/lucas77x/laucha/internal/config"
	"github.com/lucas77x/laucha/internal/i18n"
	"github.com/lucas77x/laucha/internal/launcher"
	"github.com/lucas77x/laucha/internal/search"
	"github.com/lucas77x/laucha/internal/skin"
)

const (
	defaultRowHeight = 46 // fallbacks when a skin omits [rows]
	defaultIconSize  = 30
	queryLimit       = 50 // results kept for scrolling beyond the visible rows
)

// Version is stamped by the release pipeline via -ldflags; dev builds
// keep the suffix so update checks never downgrade anyone.
var Version = "1.0.0-dev"

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

// Reconfigurer re-applies indexing settings at runtime.
type Reconfigurer interface {
	Reconfigure(roots []string, filter config.Filter)
}

// Deps groups the collaborators the bar needs; every field but Apps
// is optional. The bar builds its own engine so the Apps/Files
// switches in the settings apply live.
type Deps struct {
	Apps    search.Provider
	Files   search.Provider
	Recents RecentSource
	Usage   UsageRecorder
	Stats   search.Usage
	Reindex Reconfigurer
}

type Bar struct {
	app      fyne.App
	win      fyne.Window
	cfg      config.Config
	skin     skin.Skin
	engine   *search.Engine
	recents  RecentSource
	usage    UsageRecorder
	reindex  Reconfigurer
	input    *searchEntry
	list     *widget.List
	results  []launcher.Entry
	selected int
	resident bool // tray or hotkey active: hide instead of quitting
	visible  bool

	trayActive   bool
	hotkeyActive bool
	unbindHotkey func()

	trayMenu   *fyne.Menu
	trayToggle *fyne.MenuItem
	about      fyne.Window
	settings   fyne.Window
	border     *canvas.Rectangle
}

// Show brings the bar to the front; safe to call from any goroutine.
func (b *Bar) Show() { fyne.Do(b.show) }

func New(cfg config.Config, deps Deps) *Bar {
	appIcon := fyne.NewStaticResource("icon.svg", assets.IconSVG)
	app.SetMetadata(fyne.AppMetadata{
		ID:      "com.github.lucas77x.laucha",
		Name:    "laucha",
		Version: Version,
		Icon:    appIcon,
	})
	b := &Bar{
		app:     app.NewWithID("com.github.lucas77x.laucha"),
		cfg:     cfg,
		recents: deps.Recents,
		usage:   deps.Usage,
		reindex: deps.Reindex,
	}
	b.engine = search.NewEngine(
		search.Toggle{Provider: deps.Apps, Enabled: func() bool { return b.cfg.Search.Apps }},
		search.Toggle{Provider: deps.Files, Enabled: func() bool { return b.cfg.Search.Files }},
	)
	if deps.Stats != nil {
		b.engine.SetUsage(deps.Stats)
	}
	b.loadSkin()
	b.app.SetIcon(appIcon)
	b.win = b.newWindow()
	b.win.SetTitle("laucha")
	b.input = newSearchEntry(b.handleKey)
	b.input.SetPlaceHolder(i18n.T("Search apps and files…"))
	b.input.OnChanged = b.search
	b.list = b.newList()
	b.list.OnSelected = func(id widget.ListItemID) { b.selected = id }

	gear := widget.NewButtonWithIcon("", theme.SettingsIcon(), b.showSettings)
	gear.Importance = widget.LowImportance
	top := container.NewBorder(nil, nil, nil, gear, b.input)
	content := container.NewBorder(top, nil, nil, nil, b.list)

	b.border = canvas.NewRectangle(color.Transparent)
	b.applyBorder()
	layers := []fyne.CanvasObject{}
	if bg := b.skin.BackgroundImagePath(); bg != "" {
		image := canvas.NewImageFromFile(bg)
		image.FillMode = canvas.ImageFillStretch
		layers = append(layers, image)
	}
	layers = append(layers, container.NewPadded(content), b.border)
	b.win.SetContent(container.NewStack(layers...))
	b.resizeBar()
	b.win.CenterOnScreen()
	b.search("")
	return b
}

// loadSkin loads the configured skin, falling back to the built-in
// classic on error.
func (b *Bar) loadSkin() {
	s, err := skin.Load(b.cfg.Window.Skin)
	if err != nil {
		log.Printf("skin %q: %v (using classic)", b.cfg.Window.Skin, err)
	}
	b.skin = s
}

// applyBorder styles the bar outline from the skin.
func (b *Bar) applyBorder() {
	if b.border == nil {
		return
	}
	if c, ok := skin.ParseColor(b.skin.Border.Color); ok {
		b.border.StrokeColor = c
	} else {
		b.border.StrokeColor = color.Transparent
	}
	b.border.StrokeWidth = b.skin.Border.Width
	b.border.CornerRadius = b.skin.Border.Radius
	b.border.Refresh()
}

func (b *Bar) rowHeight() float32 {
	if b.skin.Rows.Height > 0 {
		return b.skin.Rows.Height
	}
	return defaultRowHeight
}

func (b *Bar) iconSize() float32 {
	if b.skin.Rows.IconSize > 0 {
		return b.skin.Rows.IconSize
	}
	return defaultIconSize
}

// resizeBar applies the configured width and visible rows; the extra
// padding accounts for the frame inset around the content.
func (b *Bar) resizeBar() {
	height := b.input.MinSize().Height + float32(b.cfg.Window.MaxItems)*b.rowHeight() + 8
	b.win.Resize(fyne.NewSize(b.cfg.Window.Width, height))
}

// Run shows the bar and blocks until the app exits. With a tray icon
// or a registered hotkey the app stays resident: closing the bar
// only hides it.
func (b *Bar) Run() {
	b.trayActive = b.setupTray()
	b.hotkeyActive = b.bindHotkey(b.cfg.Hotkey)
	b.resident = b.trayActive || b.hotkeyActive
	b.applyTheme()

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

	// Starting hidden only makes sense while something (tray or
	// hotkey) can bring the bar back.
	if b.cfg.Behavior.StartHidden && b.resident {
		b.visible = false
		b.refreshTrayToggle()
		b.app.Run()
		return
	}

	b.visible = true
	b.refreshTrayToggle()
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
	b.refreshTrayToggle()
}

// hide keeps the app resident when possible, otherwise quits.
func (b *Bar) hide() {
	if !b.resident {
		b.app.Quit()
		return
	}
	b.visible = false
	b.win.Hide()
	b.refreshTrayToggle()
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
			icon.SetMinSize(fyne.NewSize(b.iconSize(), b.iconSize()))
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
	if b.recents == nil || !b.cfg.Behavior.ShowRecentOnOpen || !b.cfg.Search.Files {
		return nil
	}
	return b.recents.Recent(queryLimit)
}

// applyLive applies every setting that can change at runtime; only a
// language change and disabling the tray still need a restart.
func (b *Bar) applyLive(old config.Config) {
	if b.cfg.Hotkey != old.Hotkey {
		b.rebindHotkey()
	}
	if b.cfg.Window.Width != old.Window.Width || b.cfg.Window.MaxItems != old.Window.MaxItems {
		b.resizeBar()
	}
	if b.cfg.Window.Skin != old.Window.Skin {
		b.loadSkin()
		b.applyTheme()
		b.applyBorder()
		b.resizeBar()
	}
	b.setTrayVisible(b.cfg.Behavior.ShowTrayIcon)
	b.refreshTrayToggle()
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
		argv := entry.Exec
		if entry.Terminal {
			argv = terminalCommand(argv)
		}
		cmd = exec.Command(argv[0], argv[1:]...)
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
