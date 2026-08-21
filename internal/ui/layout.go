package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

// fractionLayout sizes its single child to a fraction of the
// available width, so form fields do not stretch edge to edge.
type fractionLayout struct {
	fraction float32
	centered bool
}

func (f fractionLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.Size{}
	}
	return objects[0].MinSize()
}

func (f fractionLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}
	child := objects[0]
	width := size.Width * f.fraction
	height := child.MinSize().Height
	x := float32(0)
	if f.centered {
		x = (size.Width - width) / 2
	}
	child.Resize(fyne.NewSize(width, height))
	child.Move(fyne.NewPos(x, (size.Height-height)/2))
}

// fraction wraps a widget so it takes the given share of the row.
func fraction(object fyne.CanvasObject, share float32) fyne.CanvasObject {
	return container.New(fractionLayout{fraction: share}, object)
}

// fractionCentered is fraction with the child centered in the row.
func fractionCentered(object fyne.CanvasObject, share float32) fyne.CanvasObject {
	return container.New(fractionLayout{fraction: share, centered: true}, object)
}
