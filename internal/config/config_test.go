package config

import (
	"os"
	"testing"
)

func TestLoadCreatesDefaults(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Window.MaxItems != DefaultItems {
		t.Errorf("MaxItems = %d, want %d", cfg.Window.MaxItems, DefaultItems)
	}

	path, err := Path()
	if err != nil {
		t.Fatalf("Path: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("config file not written: %v", err)
	}
}

func TestLoadClampsMaxItems(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want int
	}{
		{"below minimum", 1, MinItems},
		{"above maximum", 42, MaxItems},
		{"in range", 7, 7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())

			cfg := Default()
			cfg.Window.MaxItems = tt.in
			if err := cfg.Save(); err != nil {
				t.Fatalf("Save: %v", err)
			}

			loaded, err := Load()
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if loaded.Window.MaxItems != tt.want {
				t.Errorf("MaxItems = %d, want %d", loaded.Window.MaxItems, tt.want)
			}
		})
	}
}
