package ui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"image/color"
)

func TestFractionLayoutSizesChild(t *testing.T) {
	child := canvas.NewRectangle(color.Black)
	child.SetMinSize(fyne.NewSize(10, 20))
	wrapper := fraction(child, 0.5)

	wrapper.Resize(fyne.NewSize(200, 40))

	if got := child.Size().Width; got != 100 {
		t.Errorf("child width = %v, want 100 (50%% of 200)", got)
	}
	if got := child.Position().X; got != 0 {
		t.Errorf("left-aligned child X = %v, want 0", got)
	}
}

func TestFractionCenteredCentersChild(t *testing.T) {
	child := canvas.NewRectangle(color.Black)
	child.SetMinSize(fyne.NewSize(10, 20))
	wrapper := fractionCentered(child, 0.5)

	wrapper.Resize(fyne.NewSize(200, 40))

	if got := child.Position().X; got != 50 {
		t.Errorf("centered child X = %v, want 50", got)
	}
}

func TestFractionMinSizeFollowsChild(t *testing.T) {
	child := canvas.NewRectangle(color.Black)
	child.SetMinSize(fyne.NewSize(30, 15))

	if got := fraction(child, 0.7).MinSize(); got.Width != 30 || got.Height != 15 {
		t.Errorf("MinSize = %v, want child's 30x15", got)
	}
}
