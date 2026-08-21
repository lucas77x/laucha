// laucha is a minimalist keyboard-driven launcher for the desktop.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/lucas77x/laucha/internal/apps"
	"github.com/lucas77x/laucha/internal/autostart"
	"github.com/lucas77x/laucha/internal/config"
	"github.com/lucas77x/laucha/internal/i18n"
	"github.com/lucas77x/laucha/internal/index"
	"github.com/lucas77x/laucha/internal/install"
	"github.com/lucas77x/laucha/internal/instance"
	"github.com/lucas77x/laucha/internal/ui"
	"github.com/lucas77x/laucha/internal/usage"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			if err := install.Install(); err != nil {
				log.Fatalf("install: %v", err)
			}
			fmt.Println("laucha installed: check your applications menu")
			return
		case "uninstall":
			if err := install.Uninstall(); err != nil {
				log.Fatalf("uninstall: %v", err)
			}
			fmt.Println("laucha removed from the applications menu")
			return
		}
	}

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

	deps := ui.Deps{Apps: apps.NewProvider()}
	if idx, err := index.New(cfg.EffectiveRoots(), cfg.EffectiveFilter()); err != nil {
		log.Printf("file index disabled: %v", err)
	} else {
		defer idx.Close()
		deps.Files = idx
		deps.Recents = idx
		deps.Reindex = idx
	}
	if opens, err := usage.Open(); err != nil {
		log.Printf("usage stats disabled: %v", err)
	} else {
		defer opens.Close()
		deps.Usage = opens
		deps.Stats = opens
	}

	bar := ui.New(cfg, deps)
	if stop, err := instance.Listen(bar.Show); err != nil {
		log.Printf("single instance: %v", err)
	} else {
		defer stop()
	}
	bar.Run()
}
