// laucha is a minimalist keyboard-driven launcher for the desktop.
package main

import (
	"log"

	"github.com/lucas77x/laucha/internal/apps"
	"github.com/lucas77x/laucha/internal/config"
	"github.com/lucas77x/laucha/internal/i18n"
	"github.com/lucas77x/laucha/internal/index"
	"github.com/lucas77x/laucha/internal/search"
	"github.com/lucas77x/laucha/internal/ui"
	"github.com/lucas77x/laucha/internal/usage"
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
	var recents ui.RecentSource
	if cfg.Search.Files {
		idx, err := index.New(cfg.Search, cfg.Filter)
		if err != nil {
			log.Printf("file index disabled: %v", err)
		} else {
			defer idx.Close()
			providers = append(providers, idx)
			recents = idx
		}
	}

	engine := search.NewEngine(providers...)
	deps := ui.Deps{Engine: engine, Recents: recents}
	if opens, err := usage.Open(); err != nil {
		log.Printf("usage stats disabled: %v", err)
	} else {
		defer opens.Close()
		engine.SetUsage(opens)
		deps.Usage = opens
	}

	ui.New(cfg, deps).Run()
}
