// laucha is a minimalist keyboard-driven launcher for the desktop.
package main

import (
	"log"

	"github.com/lucas77x/laucha/internal/apps"
	"github.com/lucas77x/laucha/internal/config"
	"github.com/lucas77x/laucha/internal/i18n"
	"github.com/lucas77x/laucha/internal/search"
	"github.com/lucas77x/laucha/internal/ui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}
	if err := i18n.Init(cfg.Language); err != nil {
		log.Printf("i18n: %v (using English)", err)
	}

	var providers []search.Provider
	if cfg.Search.Apps {
		providers = append(providers, apps.NewProvider())
	}

	ui.New(cfg, search.NewEngine(providers...)).Run()
}
