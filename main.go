// laucha is a minimalist keyboard-driven launcher for the desktop.
package main

import (
	"log"

	"github.com/lucas77x/laucha/internal/apps"
	"github.com/lucas77x/laucha/internal/autostart"
	"github.com/lucas77x/laucha/internal/config"
	"github.com/lucas77x/laucha/internal/i18n"
	"github.com/lucas77x/laucha/internal/index"
	"github.com/lucas77x/laucha/internal/instance"
	"github.com/lucas77x/laucha/internal/search"
	"github.com/lucas77x/laucha/internal/ui"
	"github.com/lucas77x/laucha/internal/usage"
)

func main() {
	if instance.NotifyRunning() {
		return // the running instance was asked to show its bar
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}
	if err := autostart.Sync(cfg.Behavior.Autostart); err != nil {
		log.Printf("autostart: %v", err)
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

	bar := ui.New(cfg, deps)
	if stop, err := instance.Listen(bar.Show); err != nil {
		log.Printf("single instance: %v", err)
	} else {
		defer stop()
	}
	bar.Run()
}
