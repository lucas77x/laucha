package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"github.com/lucas77x/laucha/internal/i18n"
)

// hotkeyCapture records a key combination: click the field, press
// the combo, done. Esc cancels and keeps the previous value. A
// combination requires at least one modifier, since a bare key would
// be a terrible global hotkey.
type hotkeyCapture struct {
	widget.Entry
	mods      map[string]bool
	previous  string
	recording bool
}

func newHotkeyCapture(current string) *hotkeyCapture {
	h := &hotkeyCapture{mods: map[string]bool{}}
	h.ExtendBaseWidget(h)
	h.SetText(current)
	return h
}

func (h *hotkeyCapture) FocusGained() {
	h.previous = h.Text
	h.recording = true
	h.mods = map[string]bool{}
	h.SetText(i18n.T("Press the shortcut…"))
	h.Entry.FocusGained()
}

func (h *hotkeyCapture) FocusLost() {
	if h.recording {
		h.recording = false
		h.SetText(h.previous)
	}
	h.Entry.FocusLost()
}

// TypedRune and TypedKey swallow input: combos come through KeyDown.
func (h *hotkeyCapture) TypedRune(rune)          {}
func (h *hotkeyCapture) TypedKey(*fyne.KeyEvent) {}
func (h *hotkeyCapture) KeyUp(key *fyne.KeyEvent) {
	if !h.recording {
		return
	}
	if token, ok := modifierTokensByKey[key.Name]; ok {
		delete(h.mods, token)
		h.SetText(h.partial())
	}
}

func (h *hotkeyCapture) KeyDown(key *fyne.KeyEvent) {
	if !h.recording {
		return
	}
	if token, ok := modifierTokensByKey[key.Name]; ok {
		h.mods[token] = true
		h.SetText(h.partial())
		return
	}
	if key.Name == fyne.KeyEscape {
		h.recording = false
		h.SetText(h.previous)
		return
	}
	token, ok := captureKeyToken(key.Name)
	if !ok || len(h.heldModifiers()) == 0 {
		return
	}
	h.recording = false
	h.SetText(strings.Join(append(h.heldModifiers(), token), "+"))
}

func (h *hotkeyCapture) partial() string {
	held := h.heldModifiers()
	if len(held) == 0 {
		return i18n.T("Press the shortcut…")
	}
	return strings.Join(held, "+") + "+…"
}

// heldModifiers returns the pressed modifiers in canonical order.
func (h *hotkeyCapture) heldModifiers() []string {
	var out []string
	for _, token := range []string{"ctrl", "shift", "alt", "super"} {
		if h.mods[token] {
			out = append(out, token)
		}
	}
	return out
}

var modifierTokensByKey = map[fyne.KeyName]string{
	desktop.KeyControlLeft:  "ctrl",
	desktop.KeyControlRight: "ctrl",
	desktop.KeyShiftLeft:    "shift",
	desktop.KeyShiftRight:   "shift",
	desktop.KeyAltLeft:      "alt",
	desktop.KeyAltRight:     "alt",
	desktop.KeySuperLeft:    "super",
	desktop.KeySuperRight:   "super",
}

// captureKeyToken maps a Fyne key to the parser's vocabulary, so a
// captured combo is always a valid hotkey spec.
func captureKeyToken(name fyne.KeyName) (string, bool) {
	token := strings.ToLower(string(name))
	_, ok := keyNames[token]
	return token, ok
}
