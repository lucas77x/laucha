package search

import "github.com/lucas77x/laucha/internal/launcher"

// Toggle wraps a provider behind an on/off switch consulted on every
// query, so enabling or disabling a source applies immediately.
type Toggle struct {
	Provider Provider // may be nil when the source is unavailable
	Enabled  func() bool
}

func (t Toggle) Entries() []launcher.Entry {
	if t.Provider == nil || !t.Enabled() {
		return nil
	}
	return t.Provider.Entries()
}
