package ui

import (
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/lucas77x/laucha/internal/autostart"
	"github.com/lucas77x/laucha/internal/config"
	"github.com/lucas77x/laucha/internal/i18n"
	"github.com/lucas77x/laucha/internal/skin"
)

// showSettings opens the Settings window, or refocuses it when
// already open. Every change saves and applies automatically; the
// status line only surfaces errors and the language restart notice.
func (b *Bar) showSettings() {
	if b.settings != nil {
		b.settings.Show()
		b.settings.RequestFocus()
		return
	}

	defaults := config.Default()
	loading := false // guards programmatic SetSelected/SetChecked

	status := widget.NewLabel("")
	var statusGen int
	showStatus := func(text string, importance widget.Importance) {
		statusGen++
		generation := statusGen
		status.Importance = importance
		status.SetText(text)
		time.AfterFunc(4*time.Second, func() {
			fyne.Do(func() {
				if generation == statusGen {
					status.SetText("")
				}
			})
		})
	}

	// commit mutates a copy of the live config, persists it and
	// applies every side effect: live settings, autostart entry and
	// reindexing when the effective search settings changed.
	commit := func(mutate func(*config.Config)) {
		old := b.cfg
		next := b.cfg
		mutate(&next)
		next.Clamp()
		if reflect.DeepEqual(old, next) {
			return
		}
		if err := next.Save(); err != nil {
			showStatus(err.Error(), widget.DangerImportance)
			return
		}
		b.cfg = next
		b.applyLive(old)
		if next.Behavior.Autostart != old.Behavior.Autostart {
			if err := autostart.Sync(next.Behavior.Autostart); err != nil {
				log.Printf("autostart: %v", err)
			}
		}
		if b.reindex != nil &&
			(!reflect.DeepEqual(old.EffectiveRoots(), next.EffectiveRoots()) ||
				!reflect.DeepEqual(old.EffectiveFilter(), next.EffectiveFilter())) {
			b.reindex.Reconfigure(next.EffectiveRoots(), next.EffectiveFilter())
		}
		if next.Language != old.Language {
			showStatus(i18n.T("Saved — some changes apply after restart"), widget.WarningImportance)
		}
	}

	// General
	language := widget.NewSelect([]string{i18n.T("System"), "English", "Español"}, nil)
	language.SetSelected(languageLabel(b.cfg.Language))
	language.OnChanged = func(selected string) {
		if !loading {
			commit(func(c *config.Config) { c.Language = languageCode(selected) })
		}
	}
	hotkey := newHotkeyCapture(b.cfg.Hotkey, b.suspendHotkey, b.resumeHotkey)
	hotkey.onCaptured = func(combo string) {
		commit(func(c *config.Config) { c.Hotkey = combo })
	}
	newCheck := func(label string, checked bool, apply func(*config.Config, bool)) *widget.Check {
		check := widget.NewCheck(label, nil)
		check.SetChecked(checked)
		check.OnChanged = func(on bool) {
			if !loading {
				commit(func(c *config.Config) { apply(c, on) })
			}
		}
		return check
	}
	autostartCheck := newCheck(i18n.T("Start at login"), b.cfg.Behavior.Autostart,
		func(c *config.Config, on bool) { c.Behavior.Autostart = on })
	startHiddenCheck := newCheck(i18n.T("Start minimized"), b.cfg.Behavior.StartHidden,
		func(c *config.Config, on bool) { c.Behavior.StartHidden = on })
	general := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem(i18n.T("Language"), fraction(language, 0.6)),
			widget.NewFormItem(i18n.T("Hotkey"), fraction(hotkey, 0.6)),
		),
		autostartCheck, startHiddenCheck,
	)

	// Behavior
	recentCheck := newCheck(i18n.T("Show recent files on open"), b.cfg.Behavior.ShowRecentOnOpen,
		func(c *config.Config, on bool) { c.Behavior.ShowRecentOnOpen = on })
	minimizeCheck := newCheck(i18n.T("Minimize on close"), b.cfg.Behavior.MinimizeOnClose,
		func(c *config.Config, on bool) { c.Behavior.MinimizeOnClose = on })
	focusCheck := newCheck(i18n.T("Hide on focus loss"), b.cfg.Behavior.HideOnFocusLost,
		func(c *config.Config, on bool) { c.Behavior.HideOnFocusLost = on })
	trayCheck := newCheck(i18n.T("Show tray icon"), b.cfg.Behavior.ShowTrayIcon,
		func(c *config.Config, on bool) { c.Behavior.ShowTrayIcon = on })
	behavior := container.NewVBox(recentCheck, minimizeCheck, focusCheck, trayCheck)

	// Search
	appsCheck := newCheck(i18n.T("Search applications"), b.cfg.Search.Apps,
		func(c *config.Config, on bool) { c.Search.Apps = on })
	filesCheck := newCheck(i18n.T("Search files"), b.cfg.Search.Files,
		func(c *config.Config, on bool) { c.Search.Files = on })

	defaultLabel := i18n.T("Default search configuration")
	advancedLabel := i18n.T("Advanced search configuration")
	excludeLabel := i18n.T("Exclude listed")
	includeLabel := i18n.T("Include only listed")

	advancedOn := b.cfg.Search.Advanced
	rootsVals := []string{}
	rootsBox := container.NewVBox()

	filterModeRadio := widget.NewRadioGroup([]string{excludeLabel, includeLabel}, nil)
	filterModeRadio.Horizontal = true
	extensionsEntry := newCommitEntry(false)
	namesEntry := newCommitEntry(false)
	patternsEntry := newCommitEntry(true)
	patternsEntry.SetMinRowsVisible(4)
	patternsEntry.Wrapping = fyne.TextWrapOff // long regexes scroll horizontally

	// commitSearch persists the whole advanced state from the
	// current field values; invalid regexes block the save.
	commitSearch := func() {
		patterns := nonEmptyLines(patternsEntry.Text)
		for _, pattern := range patterns {
			if _, err := regexp.Compile(pattern); err != nil {
				showStatus(i18n.T("Invalid pattern")+": "+pattern, widget.DangerImportance)
				return
			}
		}
		commit(func(c *config.Config) {
			c.Search.Advanced = advancedOn
			if !advancedOn {
				return
			}
			c.Search.Roots = append([]string{}, rootsVals...)
			if len(c.Search.Roots) == 0 {
				c.Search.Roots = defaults.Search.Roots
			}
			if filterModeRadio.Selected == includeLabel {
				c.Filter.Mode = "include-only"
			} else {
				c.Filter.Mode = "exclude"
			}
			c.Filter.Extensions = splitCommaList(extensionsEntry.Text)
			c.Filter.Names = splitCommaList(namesEntry.Text)
			c.Filter.Patterns = patterns
		})
	}

	var refreshRoots func()
	addRoot := widget.NewButtonWithIcon(i18n.T("Add folder"), theme.FolderNewIcon(), func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			rootsVals = append(rootsVals, uri.Path())
			refreshRoots()
			commitSearch()
		}, b.settings)
	})
	refreshRoots = func() {
		rootsBox.Objects = nil
		for i, root := range rootsVals {
			at := i
			remove := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
				rootsVals = append(rootsVals[:at], rootsVals[at+1:]...)
				refreshRoots()
				commitSearch()
			})
			rootsBox.Add(container.NewBorder(nil, nil, nil, remove, widget.NewLabel(root)))
		}
		rootsBox.Refresh()
	}

	populate := func(src config.Config) {
		loading = true
		rootsVals = append([]string{}, src.Search.Roots...)
		extensionsEntry.SetText(strings.Join(src.Filter.Extensions, ", "))
		namesEntry.SetText(strings.Join(src.Filter.Names, ", "))
		patternsEntry.SetText(strings.Join(src.Filter.Patterns, "\n"))
		if src.Filter.Mode == "include-only" {
			filterModeRadio.SetSelected(includeLabel)
		} else {
			filterModeRadio.SetSelected(excludeLabel)
		}
		refreshRoots()
		loading = false
	}

	filterModeRadio.OnChanged = func(string) {
		if !loading {
			commitSearch()
		}
	}
	extensionsEntry.onCommit = func(string) { commitSearch() }
	namesEntry.onCommit = func(string) { commitSearch() }
	patternsEntry.onCommit = func(string) { commitSearch() }

	restore := widget.NewButton(i18n.T("Restore defaults"), func() {
		populate(defaults)
		commitSearch()
	})

	rootsBg := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	rootsBg.CornerRadius = 8
	rootsArea := fractionCentered(container.NewStack(rootsBg, container.NewPadded(rootsBox)), 0.7)

	advNote := widget.NewLabel(i18n.T("Advanced settings replace the defaults"))
	advNote.TextStyle = fyne.TextStyle{Italic: true}

	advancedBox := container.NewVBox(
		widget.NewSeparator(),
		widget.NewLabel(i18n.T("Indexed folders")),
		rootsArea,
		container.NewCenter(addRoot),
		widget.NewForm(
			widget.NewFormItem(i18n.T("Filter mode"), filterModeRadio),
			widget.NewFormItem(i18n.T("Extensions"), fraction(extensionsEntry, 0.6)),
			widget.NewFormItem(i18n.T("Names"), fraction(namesEntry, 0.6)),
		),
		widget.NewLabel(i18n.T("Patterns (regex, one per line)")),
		fraction(patternsEntry, 0.7),
		container.NewCenter(restore),
	)

	setAdvanced := func(on bool) {
		advancedOn = on
		if on {
			advancedBox.Show()
		} else {
			advancedBox.Hide()
		}
	}

	configRadio := widget.NewRadioGroup([]string{defaultLabel, advancedLabel}, nil)
	loading = true
	if b.cfg.Search.Advanced {
		populate(b.cfg)
		configRadio.SetSelected(advancedLabel)
	} else {
		populate(defaults)
		configRadio.SetSelected(defaultLabel)
	}
	loading = false
	setAdvanced(b.cfg.Search.Advanced)
	configRadio.OnChanged = func(selected string) {
		if loading {
			return
		}
		on := selected == advancedLabel
		if on && b.cfg.Search.Advanced {
			populate(b.cfg) // bring back the saved customization
		}
		if !on {
			populate(defaults)
		}
		setAdvanced(on)
		commitSearch()
	}

	searchTab := container.NewScroll(container.NewVBox(
		appsCheck, filesCheck,
		widget.NewSeparator(),
		configRadio,
		advNote,
		advancedBox,
	))

	// Display
	width := newCommitEntry(false)
	width.SetText(strconv.Itoa(int(b.cfg.Window.Width)))
	width.onCommit = func(text string) {
		v, err := strconv.Atoi(strings.TrimSpace(text))
		if err != nil {
			width.SetText(strconv.Itoa(int(b.cfg.Window.Width)))
			return
		}
		commit(func(c *config.Config) { c.Window.Width = float32(v) })
		width.SetText(strconv.Itoa(int(b.cfg.Window.Width))) // reflect clamping
	}
	items := widget.NewSelect([]string{"3", "4", "5", "6", "7", "8", "9", "10"}, nil)
	items.SetSelected(strconv.Itoa(b.cfg.Window.MaxItems))
	items.OnChanged = func(selected string) {
		if loading {
			return
		}
		if v, err := strconv.Atoi(selected); err == nil {
			commit(func(c *config.Config) { c.Window.MaxItems = v })
		}
	}
	skinSelect := widget.NewSelect(skin.Available(), nil)
	skinSelect.SetSelected(b.cfg.Window.Skin)
	skinSelect.OnChanged = func(selected string) {
		if !loading {
			commit(func(c *config.Config) { c.Window.Skin = selected })
		}
	}
	display := container.NewVBox(widget.NewForm(
		widget.NewFormItem(i18n.T("Window width"), fraction(width, 0.4)),
		widget.NewFormItem(i18n.T("Visible items"), fraction(items, 0.4)),
		widget.NewFormItem(i18n.T("Skin"), fraction(skinSelect, 0.6)),
	))

	tabs := container.NewAppTabs(
		container.NewTabItem(i18n.T("General"), general),
		container.NewTabItem(i18n.T("Behavior"), behavior),
		container.NewTabItem(i18n.T("Search"), searchTab),
		container.NewTabItem(i18n.T("Display"), display),
		container.NewTabItem(i18n.T("About"), b.aboutContent()),
	)
	tabs.SetTabLocation(container.TabLocationLeading)

	w := b.app.NewWindow(i18n.T("Settings"))
	w.SetContent(container.NewBorder(nil, status, nil, nil, tabs))
	w.Resize(fyne.NewSize(680, 500))
	w.CenterOnScreen()
	w.SetOnClosed(func() { b.settings = nil })
	b.settings = w
	w.Show()
}

func splitCommaList(s string) []string {
	out := []string{}
	for _, part := range strings.Split(s, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func nonEmptyLines(s string) []string {
	out := []string{}
	for _, line := range strings.Split(s, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func languageLabel(code string) string {
	switch code {
	case "en":
		return "English"
	case "es":
		return "Español"
	default:
		return i18n.T("System")
	}
}

func languageCode(label string) string {
	switch label {
	case "English":
		return "en"
	case "Español":
		return "es"
	default:
		return "system"
	}
}
