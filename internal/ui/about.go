package ui

import (
	"errors"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/lucas77x/laucha/assets"
	"github.com/lucas77x/laucha/internal/i18n"
	"github.com/lucas77x/laucha/internal/update"
)

const repoURL = "https://github.com/lucas77x/laucha"

// aboutContent builds the shared About view used by the About window
// and the Settings tab.
func (b *Bar) aboutContent() fyne.CanvasObject {
	icon := canvas.NewImageFromResource(fyne.NewStaticResource("icon.svg", assets.IconSVG))
	icon.FillMode = canvas.ImageFillContain
	icon.SetMinSize(fyne.NewSize(96, 96))

	title := widget.NewLabelWithStyle("laucha "+b.app.Metadata().Version,
		fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	subtitle := widget.NewLabelWithStyle(i18n.T("A minimalist launcher"),
		fyne.TextAlignCenter, fyne.TextStyle{})

	repo, _ := url.Parse(repoURL)
	link := widget.NewHyperlink("github.com/lucas77x/laucha", repo)

	updateStatus := widget.NewLabel("")
	updateStatus.Alignment = fyne.TextAlignCenter
	checkUpdates := widget.NewButton(i18n.T("Check for updates"), func() {
		updateStatus.SetText(i18n.T("Checking…"))
		go func() {
			latest, err := update.CheckLatest()
			fyne.Do(func() {
				switch {
				case errors.Is(err, update.ErrNoReleases):
					updateStatus.SetText(i18n.T("No releases published yet"))
				case err != nil:
					updateStatus.SetText(i18n.T("Could not check for updates"))
				case update.IsNewer(b.app.Metadata().Version, latest):
					updateStatus.SetText(i18n.T("New version available") + ": " + latest)
					if page, err := url.Parse(update.ReleasePage); err == nil {
						_ = b.app.OpenURL(page)
					}
				default:
					updateStatus.SetText(i18n.T("You are up to date"))
				}
			})
		}()
	})

	return container.NewVBox(icon, title, subtitle, container.NewCenter(link),
		container.NewCenter(checkUpdates), updateStatus)
}

// showAbout opens the About window, or refocuses it when already
// open.
func (b *Bar) showAbout() {
	if b.about != nil {
		b.about.Show()
		b.about.RequestFocus()
		return
	}

	w := b.app.NewWindow(i18n.T("About") + " laucha")
	w.SetContent(b.aboutContent())
	w.Resize(fyne.NewSize(320, 240))
	w.CenterOnScreen()
	w.SetOnClosed(func() { b.about = nil })
	b.about = w
	w.Show()
}
