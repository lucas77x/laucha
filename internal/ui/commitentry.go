package ui

import "fyne.io/fyne/v2/widget"

// commitEntry saves its content when the user finishes editing —
// Enter or focus loss — never on every keystroke, so half-typed
// values and broken regexes are never persisted.
type commitEntry struct {
	widget.Entry
	onCommit func(string)
}

func newCommitEntry(multiline bool) *commitEntry {
	e := &commitEntry{}
	e.MultiLine = multiline
	e.ExtendBaseWidget(e)
	e.OnSubmitted = func(text string) {
		if e.onCommit != nil {
			e.onCommit(text)
		}
	}
	return e
}

func (e *commitEntry) FocusLost() {
	e.Entry.FocusLost()
	if e.onCommit != nil {
		e.onCommit(e.Text)
	}
}
