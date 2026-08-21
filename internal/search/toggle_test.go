package search

import "testing"

func TestToggleSwitchesLive(t *testing.T) {
	enabled := true
	toggle := Toggle{Provider: provider("Spotify"), Enabled: func() bool { return enabled }}

	if got := len(toggle.Entries()); got != 1 {
		t.Errorf("enabled Entries len = %d, want 1", got)
	}
	enabled = false
	if got := toggle.Entries(); got != nil {
		t.Errorf("disabled Entries = %v, want nil", got)
	}
}

func TestToggleWithNilProvider(t *testing.T) {
	toggle := Toggle{Enabled: func() bool { return true }}

	if got := toggle.Entries(); got != nil {
		t.Errorf("nil provider Entries = %v, want nil", got)
	}
}
