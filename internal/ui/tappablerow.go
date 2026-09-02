package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// tappableRow makes a list row open on click regardless of selection
// state — widget.List only reports selection CHANGES, so clicking the
// already-selected first row would otherwise do nothing. It also gives
// the row the height the skin asks for, so the window fits exactly the
// configured number of rows.
type tappableRow struct {
	widget.BaseWidget
	content   fyne.CanvasObject
	minHeight float32
	onTap     func()
}

func newTappableRow(content fyne.CanvasObject, minHeight float32) *tappableRow {
	r := &tappableRow{content: content, minHeight: minHeight}
	r.ExtendBaseWidget(r)
	return r
}

func (r *tappableRow) MinSize() fyne.Size {
	size := r.content.MinSize()
	if size.Height < r.minHeight {
		size.Height = r.minHeight
	}
	return size
}

func (r *tappableRow) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(r.content)
}

func (r *tappableRow) Tapped(*fyne.PointEvent) {
	if r.onTap != nil {
		r.onTap()
	}
}
