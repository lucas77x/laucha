package ui

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/lucas77x/laucha/assets"
	"github.com/lucas77x/laucha/internal/i18n"
)

const repoURL = "https://github.com/lucas77x/laucha"

// showAbout opens the About window, or refocuses it when already
// open.
func (b *Bar) showAbout() {
	if b.about != nil {
		b.about.Show()
		b.about.RequestFocus()
		return
	}

	icon := canvas.NewImageFromResource(fyne.NewStaticResource("icon.svg", assets.IconSVG))
	icon.FillMode = canvas.ImageFillContain
	icon.SetMinSize(fyne.NewSize(96, 96))

	title := widget.NewLabelWithStyle("laucha "+b.app.Metadata().Version,
		fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	subtitle := widget.NewLabelWithStyle(i18n.T("A minimalist keyboard-driven launcher"),
		fyne.TextAlignCenter, fyne.TextStyle{})

	repo, _ := url.Parse(repoURL)
	link := widget.NewHyperlink("github.com/lucas77x/laucha", repo)

	w := b.app.NewWindow(i18n.T("About") + " laucha")
	w.SetContent(container.NewVBox(icon, title, subtitle, container.NewCenter(link)))
	w.Resize(fyne.NewSize(320, 240))
	w.CenterOnScreen()
	w.SetOnClosed(func() { b.about = nil })
	b.about = w
	w.Show()
}
