package ui

import (
	"log"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/lucas77x/laucha/internal/autostart"
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

	// General
	language := widget.NewSelect([]string{i18n.T("System"), "English", "Español"}, nil)
	language.SetSelected(languageLabel(cfg.Language))
	hotkey := widget.NewEntry()
	hotkey.SetText(cfg.Hotkey)
	autostartCheck := widget.NewCheck(i18n.T("Start at login"), nil)
	autostartCheck.SetChecked(cfg.Behavior.Autostart)
	appsCheck := widget.NewCheck(i18n.T("Search applications"), nil)
	appsCheck.SetChecked(cfg.Search.Apps)
	filesCheck := widget.NewCheck(i18n.T("Search files"), nil)
	filesCheck.SetChecked(cfg.Search.Files)
	general := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem(i18n.T("Language"), fraction(language, 0.6)),
			widget.NewFormItem(i18n.T("Hotkey"), fraction(hotkey, 0.6)),
		),
		autostartCheck, appsCheck, filesCheck,
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
	save := widget.NewButton(i18n.T("Save"), func() {
		if _, _, err := parseHotkey(hotkey.Text); err != nil {
			showStatus(i18n.T("Invalid hotkey"), widget.DangerImportance)
			return
		}
		old := b.cfg
		cfg.Language = languageCode(language.Selected)
		cfg.Hotkey = hotkey.Text
		cfg.Behavior.Autostart = autostartCheck.Checked
		cfg.Search.Apps = appsCheck.Checked
		cfg.Search.Files = filesCheck.Checked
		cfg.Behavior.ShowRecentOnOpen = recentCheck.Checked
		cfg.Behavior.MinimizeOnClose = minimizeCheck.Checked
		cfg.Behavior.HideOnFocusLost = focusCheck.Checked
		cfg.Behavior.ShowTrayIcon = trayCheck.Checked
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

		if cfg.Language != old.Language {
			showStatus("✅ "+i18n.T("Saved — some changes apply after restart"), widget.WarningImportance)
		} else {
			showStatus("✅ "+i18n.T("Saved"), widget.SuccessImportance)
		}
	})

	w := b.app.NewWindow(i18n.T("Settings"))
	w.SetContent(container.NewBorder(nil, container.NewVBox(status, save), nil, nil, tabs))
	w.Resize(fyne.NewSize(540, 420))
	w.CenterOnScreen()
	w.SetOnClosed(func() { b.settings = nil })
	b.settings = w
	w.Show()
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
