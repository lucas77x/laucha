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
// already open. Vertical tabs keep room to grow.
func (b *Bar) showSettings() {
	if b.settings != nil {
		b.settings.Show()
		b.settings.RequestFocus()
		return
	}

	cfg := b.cfg // working copy; written back on Save
	defaults := config.Default()

	// General
	language := widget.NewSelect([]string{i18n.T("System"), "English", "Español"}, nil)
	language.SetSelected(languageLabel(cfg.Language))
	hotkey := newHotkeyCapture(cfg.Hotkey, b.suspendHotkey, b.resumeHotkey)
	autostartCheck := widget.NewCheck(i18n.T("Start at login"), nil)
	autostartCheck.SetChecked(cfg.Behavior.Autostart)
	startHiddenCheck := widget.NewCheck(i18n.T("Start minimized"), nil)
	startHiddenCheck.SetChecked(cfg.Behavior.StartHidden)
	general := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem(i18n.T("Language"), fraction(language, 0.6)),
			widget.NewFormItem(i18n.T("Hotkey"), fraction(hotkey, 0.6)),
		),
		autostartCheck, startHiddenCheck,
	)

	// Behavior
	recentCheck := widget.NewCheck(i18n.T("Show recent files on open"), nil)
	recentCheck.SetChecked(cfg.Behavior.ShowRecentOnOpen)
	minimizeCheck := widget.NewCheck(i18n.T("Minimize on close"), nil)
	minimizeCheck.SetChecked(cfg.Behavior.MinimizeOnClose)
	focusCheck := widget.NewCheck(i18n.T("Hide on focus loss"), nil)
	focusCheck.SetChecked(cfg.Behavior.HideOnFocusLost)
	trayCheck := widget.NewCheck(i18n.T("Show tray icon"), b.setTrayVisible)
	trayCheck.SetChecked(cfg.Behavior.ShowTrayIcon)
	behavior := container.NewVBox(recentCheck, minimizeCheck, focusCheck, trayCheck)

	// Search
	appsCheck := widget.NewCheck(i18n.T("Search applications"), nil)
	appsCheck.SetChecked(cfg.Search.Apps)
	filesCheck := widget.NewCheck(i18n.T("Search files"), nil)
	filesCheck.SetChecked(cfg.Search.Files)

	defaultLabel := i18n.T("Default search configuration")
	advancedLabel := i18n.T("Advanced search configuration")
	excludeLabel := i18n.T("Exclude listed")
	includeLabel := i18n.T("Include only listed")

	advancedOn := cfg.Search.Advanced
	rootsVals := []string{}
	rootsBox := container.NewVBox()

	filterModeRadio := widget.NewRadioGroup([]string{excludeLabel, includeLabel}, nil)
	filterModeRadio.Horizontal = true
	extensionsEntry := widget.NewEntry()
	namesEntry := widget.NewEntry()
	patternsEntry := widget.NewMultiLineEntry()
	patternsEntry.SetMinRowsVisible(4)
	patternsEntry.Wrapping = fyne.TextWrapOff // long regexes scroll horizontally

	var refreshRoots func()
	addRoot := widget.NewButtonWithIcon(i18n.T("Add folder"), theme.FolderNewIcon(), func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			rootsVals = append(rootsVals, uri.Path())
			refreshRoots()
		}, b.settings)
	})
	refreshRoots = func() {
		rootsBox.Objects = nil
		for i, root := range rootsVals {
			at := i
			remove := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
				rootsVals = append(rootsVals[:at], rootsVals[at+1:]...)
				refreshRoots()
			})
			rootsBox.Add(container.NewBorder(nil, nil, nil, remove, widget.NewLabel(root)))
		}
		rootsBox.Refresh()
	}

	populate := func(src config.Config) {
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
	}

	restore := widget.NewButton(i18n.T("Restore defaults"), func() {
		populate(defaults)
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
	if cfg.Search.Advanced {
		populate(cfg)
		configRadio.SetSelected(advancedLabel)
	} else {
		populate(defaults)
		configRadio.SetSelected(defaultLabel)
	}
	setAdvanced(cfg.Search.Advanced)
	configRadio.OnChanged = func(selected string) {
		on := selected == advancedLabel
		if on && cfg.Search.Advanced {
			populate(cfg) // bring back the saved customization
		}
		if !on {
			populate(defaults)
		}
		setAdvanced(on)
	}

	searchTab := container.NewScroll(container.NewVBox(
		appsCheck, filesCheck,
		widget.NewSeparator(),
		configRadio,
		advNote,
		advancedBox,
	))

	// Display
	width := widget.NewEntry()
	width.SetText(strconv.Itoa(int(cfg.Window.Width)))
	items := widget.NewSelect([]string{"3", "4", "5", "6", "7", "8", "9", "10"}, nil)
	items.SetSelected(strconv.Itoa(cfg.Window.MaxItems))
	skinSelect := widget.NewSelect(skin.Available(), nil)
	skinSelect.SetSelected(cfg.Window.Skin)
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
	// A captured combo saves itself quietly: no Save button, no
	// status message — the field just shows the new shortcut.
	hotkey.onCaptured = func(combo string) {
		cfg.Hotkey = combo
		b.cfg.Hotkey = combo
		if err := b.cfg.Save(); err != nil {
			showStatus(err.Error(), widget.DangerImportance)
		}
	}

	save := widget.NewButton(i18n.T("Save"), func() {
		if _, _, err := parseHotkey(hotkey.Text); err != nil {
			showStatus(i18n.T("Invalid hotkey"), widget.DangerImportance)
			return
		}
		old := b.cfg
		cfg.Language = languageCode(language.Selected)
		cfg.Hotkey = hotkey.Text
		cfg.Behavior.Autostart = autostartCheck.Checked
		cfg.Behavior.StartHidden = startHiddenCheck.Checked
		cfg.Search.Apps = appsCheck.Checked
		cfg.Search.Files = filesCheck.Checked
		cfg.Behavior.ShowRecentOnOpen = recentCheck.Checked
		cfg.Behavior.MinimizeOnClose = minimizeCheck.Checked
		cfg.Behavior.HideOnFocusLost = focusCheck.Checked
		cfg.Behavior.ShowTrayIcon = trayCheck.Checked
		cfg.Search.Advanced = advancedOn
		if advancedOn {
			patterns := nonEmptyLines(patternsEntry.Text)
			for _, pattern := range patterns {
				if _, err := regexp.Compile(pattern); err != nil {
					showStatus(i18n.T("Invalid pattern")+": "+pattern, widget.DangerImportance)
					return
				}
			}
			cfg.Search.Roots = append([]string{}, rootsVals...)
			if len(cfg.Search.Roots) == 0 {
				cfg.Search.Roots = defaults.Search.Roots
			}
			if filterModeRadio.Selected == includeLabel {
				cfg.Filter.Mode = "include-only"
			} else {
				cfg.Filter.Mode = "exclude"
			}
			cfg.Filter.Extensions = splitCommaList(extensionsEntry.Text)
			cfg.Filter.Names = splitCommaList(namesEntry.Text)
			cfg.Filter.Patterns = patterns
		}
		if v, err := strconv.Atoi(width.Text); err == nil {
			cfg.Window.Width = float32(v)
		}
		if v, err := strconv.Atoi(items.Selected); err == nil {
			cfg.Window.MaxItems = v
		}
		cfg.Window.Skin = skinSelect.Selected
		cfg.Clamp()

		if err := cfg.Save(); err != nil {
			showStatus(err.Error(), widget.DangerImportance)
			return
		}
		if err := autostart.Sync(cfg.Behavior.Autostart); err != nil {
			log.Printf("autostart: %v", err)
		}
		b.cfg = cfg
		b.applyLive(old)
		width.SetText(strconv.Itoa(int(cfg.Window.Width)))

		reindexing := false
		if b.reindex != nil &&
			(!reflect.DeepEqual(old.EffectiveRoots(), cfg.EffectiveRoots()) ||
				!reflect.DeepEqual(old.EffectiveFilter(), cfg.EffectiveFilter())) {
			b.reindex.Reconfigure(cfg.EffectiveRoots(), cfg.EffectiveFilter())
			reindexing = true
		}

		switch {
		case cfg.Language != old.Language:
			showStatus("✅ "+i18n.T("Saved — some changes apply after restart"), widget.WarningImportance)
		case reindexing:
			showStatus("✅ "+i18n.T("Saved — reindexing in the background"), widget.SuccessImportance)
		default:
			showStatus("✅ "+i18n.T("Saved"), widget.SuccessImportance)
		}
	})

	w := b.app.NewWindow(i18n.T("Settings"))
	w.SetContent(container.NewBorder(nil, container.NewVBox(status, save), nil, nil, tabs))
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
