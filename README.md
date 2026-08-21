# laucha

A minimalist, keyboard-driven launcher for Linux, built with Go and [Fyne](https://fyne.io). Open source, skinnable, translatable.

> **Status**: early development. The app-search MVP works; see the roadmap below.

## Why

Existing launchers kept forgetting where things are. laucha keeps a persistent index that stays fresh in real time, so a file you downloaded two seconds ago is already searchable.

## Features

- App search by icon + name — fuzzy and case-insensitive (`spo` finds Spotify, `calc` finds the calculator)
- Multi-term search across name and path, in any order: `not nextcl` (or `nextcl not`) narrows `notas.txt` down to the copy under `~/Nextcloud`
- Frecency ranking: opens are counted, so equally good matches surface what you actually use first
- Live file index: SQLite-backed and kept fresh by inotify watchers — a file downloaded seconds ago is already searchable
- Recent files, newest first, shown when the bar opens (configurable)
- Bundled file-type icons (text, pdf, spreadsheet, image, audio, video, archive, code, …)
- File results show the name plus its location relative to home
- Keyboard-first: type, navigate with arrows, launch with Enter, hide with Esc
- Resident app: system tray icon and a configurable global hotkey (`ctrl+space` by default) toggle the bar instantly; launching or losing focus hides it instead of quitting
- Single instance: launching the binary again just shows the running bar
- Settings window (tray → Settings) with vertical tabs: General / Behavior / Display / About
- Optional autostart at login, managed from Settings
- Config file with sane defaults, created on first run
- Translations: English (default) and Spanish

The first run walks the configured roots to build the index; later runs search instantly on the stored index while a background walk reconciles it. If some directories exceed the inotify watch limit, raise `fs.inotify.max_user_watches`.

If another application already owns the configured hotkey, laucha logs it and keeps running without the global shortcut.

`third_party/systray` is a vendored copy of [fyne.io/systray](https://github.com/fyne-io/systray) (Apache-2.0) with one added function, `SetIconName`: ayatana-based tray hosts (MATE, XFCE) ignore `IconPixmap`, so laucha also publishes its tray icon as a file path, the same strategy Qt applications use.

## Roadmap

- Skin engine — template-based; `classic` is the reference skin
- Self-update from GitHub Releases
- Windows and macOS support

## Requirements

- Go 1.27+
- Fyne build dependencies on Debian/Ubuntu:

```sh
sudo apt install pkg-config libgl1-mesa-dev xorg-dev libxkbcommon-dev
```

## Build & run

```sh
go build .
./laucha
```

## Configuration

`~/.config/laucha/config.toml`, created with defaults on first run:

```toml
language = "system" # or "en", "es"
hotkey = "ctrl+space"

[window]
width = 640
max_items = 4 # visible rows before scrolling (3-10)
skin = "default-dark" # or "default-light", or any drop-in skin

[behavior]
show_tray_icon = true
minimize_on_close = true
hide_on_focus_lost = true
show_recent_on_open = true
autostart = false

[search]
apps = true
files = true
roots = ["~"]

[filter]
mode = "exclude" # exclude | include-only
extensions = []
names = ["node_modules"]
patterns = ['(^|/)\.[^/]+'] # RE2 regex; default hides dotfiles
```

## Skins

A skin is a folder dropped into `skins/<name>/` next to the binary, or into `~/.config/laucha/skins/`. Each skin declares one of the predefined layout templates in its `skin.toml` and dresses it. Unset values fall back to the built-in classic look, so a three-line skin is valid. Switching skins applies live.

```toml
name = "My skin"
template = "classic"

[colors]
background = "#1B191F"       # bar and windows
foreground = "#E8E4DE"       # primary text
muted = "#8A879B"            # file paths, placeholders
accent = "#E8A0B4"           # focus, links, highlights
selection = "#3A2E36"        # selected row
input_background = "#141317" # search input

[font]
size = 15

[rows]
height = 46
icon_size = 30

[border]
color = "#E8A0B455" # bar outline; supports #RRGGBBAA
width = 1
radius = 10

[images]
background = "background.png" # optional, stretched to the window
```

Elevation, separators, hover and scrollbar colors are derived from the declared palette automatically. `skins/classic` is the reference skin.

## Translations

One JSON file per language in `internal/i18n/translations/`; keys are the English source strings. Adding a language is adding one file — contributions welcome.

## Project layout

```
main.go              composition root
internal/launcher    domain types (Entry)
internal/config      TOML settings, defaults, clamping
internal/search      ranking engine over pluggable providers
internal/apps        .desktop discovery + icon resolution
internal/i18n        translation wrapper (Fyne lang)
internal/ui          the launcher bar (Fyne)
skins/               drop-in skins
```

## License

[MIT](LICENSE)
