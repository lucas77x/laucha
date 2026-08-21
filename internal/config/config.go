// Package config loads and persists user settings as a TOML file
// under the OS config directory (~/.config/laucha/config.toml on Linux).
package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

const (
	MinItems     = 3
	MaxItems     = 10
	DefaultItems = 4

	minWidth     = 320
	defaultWidth = 640
)

type Config struct {
	Language string   `toml:"language"` // "system" or a BCP-47 code such as "en", "es"
	Hotkey   string   `toml:"hotkey"`
	Window   Window   `toml:"window"`
	Behavior Behavior `toml:"behavior"`
	Search   Search   `toml:"search"`
	Filter   Filter   `toml:"filter"`
}

type Window struct {
	Width    float32 `toml:"width"`
	MaxItems int     `toml:"max_items"` // visible rows before scrolling
	Skin     string  `toml:"skin"`
}

type Behavior struct {
	ShowTrayIcon     bool `toml:"show_tray_icon"`
	MinimizeOnClose  bool `toml:"minimize_on_close"`
	HideOnFocusLost  bool `toml:"hide_on_focus_lost"`
	ShowRecentOnOpen bool `toml:"show_recent_on_open"`
	Autostart        bool `toml:"autostart"`
	StartHidden      bool `toml:"start_hidden"` // start resident without showing the bar
}

type Search struct {
	Apps     bool     `toml:"apps"`
	Files    bool     `toml:"files"`
	Advanced bool     `toml:"advanced"` // custom roots/filter override the defaults
	Roots    []string `toml:"roots"`
}

// Filter controls what the file indexer skips (mode "exclude") or the
// only things it accepts (mode "include-only").
type Filter struct {
	Mode       string   `toml:"mode"` // exclude | include-only
	Extensions []string `toml:"extensions"`
	Names      []string `toml:"names"`
	Patterns   []string `toml:"patterns"` // RE2 regular expressions
}

func Default() Config {
	return Config{
		Language: "system",
		Hotkey:   "ctrl+space",
		Window: Window{
			Width:    defaultWidth,
			MaxItems: DefaultItems,
			Skin:     "default-dark",
		},
		Behavior: Behavior{
			ShowTrayIcon:     true,
			MinimizeOnClose:  true,
			HideOnFocusLost:  true,
			ShowRecentOnOpen: true,
		},
		Search: Search{
			Apps:  true,
			Files: true,
			Roots: []string{"~"},
		},
		Filter: Filter{
			Mode:       "exclude",
			Extensions: []string{},
			Names:      []string{"node_modules", "__pycache__"},
			Patterns: []string{
				`(^|/)\.[^/]+`,     // hidden files and directories
				`(^|/)go/pkg(/|$)`, // Go module cache
				`(^|/)snap(/|$)`,   // snap application data
			},
		},
	}
}

// EffectiveRoots returns the indexed roots: the built-in defaults
// unless the advanced search configuration is enabled. Keeping the
// defaults out of the user file lets future versions improve them
// for everyone still on the default configuration.
func (c Config) EffectiveRoots() []string {
	if c.Search.Advanced {
		return c.Search.Roots
	}
	return Default().Search.Roots
}

// EffectiveFilter returns the active filter under the same rule.
func (c Config) EffectiveFilter() Filter {
	if c.Search.Advanced {
		return c.Filter
	}
	return Default().Filter
}

// Path returns the config file location, creating its directory.
func Path() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "laucha")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

// Load reads the config file, writing the defaults first if it does
// not exist yet. Out-of-range values are clamped, never rejected.
func Load() (Config, error) {
	path, err := Path()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		cfg := Default()
		return cfg, cfg.Save()
	}
	if err != nil {
		return Config{}, err
	}
	cfg := Default()
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	cfg.Clamp()
	return cfg, nil
}

func (c Config) Save() error {
	path, err := Path()
	if err != nil {
		return err
	}
	data, err := toml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Clamp forces out-of-range values back into their valid ranges.
func (c *Config) Clamp() {
	if c.Window.MaxItems < MinItems {
		c.Window.MaxItems = MinItems
	}
	if c.Window.MaxItems > MaxItems {
		c.Window.MaxItems = MaxItems
	}
	if c.Window.Width < minWidth {
		c.Window.Width = minWidth
	}
}
