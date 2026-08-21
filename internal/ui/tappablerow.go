package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// tappableRow makes a list row open on click regardless of selection
// state — widget.List only reports selection CHANGES, so clicking the
// already-selected first row would otherwise do nothing.
type tappableRow struct {
	widget.BaseWidget
	content fyne.CanvasObject
	onTap   func()
}

func newTappableRow(content fyne.CanvasObject) *tappableRow {
	r := &tappableRow{content: content}
	r.ExtendBaseWidget(r)
	return r
}

func (r *tappableRow) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(r.content)
}

func (r *tappableRow) Tapped(*fyne.PointEvent) {
	if r.onTap != nil {
		r.onTap()
	}
}
