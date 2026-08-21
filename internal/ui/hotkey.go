package ui

import (
	"fmt"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"golang.design/x/hotkey"
)

// registerHotkey binds the configured global shortcut. Events arrive
// on a background goroutine, so UI work is bridged through fyne.Do.
func (b *Bar) registerHotkey() bool {
	mods, key, err := parseHotkey(b.cfg.Hotkey)
	if err != nil {
		log.Printf("hotkey %q disabled: %v", b.cfg.Hotkey, err)
		return false
	}
	hk := hotkey.New(mods, key)
	if err := hk.Register(); err != nil {
		log.Printf("hotkey %q not registered: %v", b.cfg.Hotkey, err)
		return false
	}
	go func() {
		for range hk.Keydown() {
			fyne.Do(b.toggle)
		}
	}()
	return true
}

// parseHotkey turns a spec such as "ctrl+space" or "super+shift+p"
// into the modifiers and key the hotkey library expects.
func parseHotkey(spec string) ([]hotkey.Modifier, hotkey.Key, error) {
	parts := strings.Split(strings.ToLower(strings.ReplaceAll(spec, " ", "")), "+")
	if len(parts) < 2 {
		return nil, 0, fmt.Errorf("expected modifier+key, got %q", spec)
	}
	var mods []hotkey.Modifier
	for _, part := range parts[:len(parts)-1] {
		mod, ok := modifierNames[part]
		if !ok {
			return nil, 0, fmt.Errorf("unknown modifier %q", part)
		}
		mods = append(mods, mod)
	}
	key, ok := keyNames[parts[len(parts)-1]]
	if !ok {
		return nil, 0, fmt.Errorf("unknown key %q", parts[len(parts)-1])
	}
	return mods, key, nil
}

var modifierNames = map[string]hotkey.Modifier{
	"ctrl":  hotkey.ModCtrl,
	"shift": hotkey.ModShift,
	"alt":   hotkey.Mod1,
	"super": hotkey.Mod4,
}

var keyNames = buildKeyNames()

func buildKeyNames() map[string]hotkey.Key {
	names := map[string]hotkey.Key{
		"space":  hotkey.KeySpace,
		"return": hotkey.KeyReturn,
		"escape": hotkey.KeyEscape,
		"tab":    hotkey.KeyTab,
	}
	for r := 'a'; r <= 'z'; r++ {
		names[string(r)] = hotkey.KeyA + hotkey.Key(r-'a')
	}
	for r := '0'; r <= '9'; r++ {
		names[string(r)] = hotkey.Key0 + hotkey.Key(r-'0')
	}
	for i := 0; i < 12; i++ {
		names[fmt.Sprintf("f%d", i+1)] = hotkey.KeyF1 + hotkey.Key(i)
	}
	return names
}
