# Contributing to laucha

Thanks for wanting to help! There are three ways to contribute, and two of them require no Go at all.

## Skins (no code needed)

A skin is a folder with a `skin.toml`. Copy `skins/default-dark/` as a starting point, rename the folder, and edit the palette. Drop it into `skins/` next to the binary or `~/.config/laucha/skins/` and it appears in Settings immediately.

- Declare only what you change — unset values fall back to the defaults.
- `on_accent` is derived by auto-contrast; override it only if the result reads badly.
- An optional `background.png` inside the folder is stretched to the window.
- To share a skin, open a pull request adding your folder under `skins/`.

See the full schema in the [README](README.md#skins).

## Translations (no code needed)

One JSON file per language in `internal/i18n/translations/`. Keys are the English source strings:

```json
{
  "Search apps and files…": "Rechercher des applications et fichiers…"
}
```

Copy `es.json`, rename it to your BCP-47 code (`fr.json`, `pt.json`, `de.json`…), translate the values, open a pull request. Untranslated keys fall back to English, so partial translations are welcome.

## Code

```sh
# Build dependencies (Debian/Ubuntu)
sudo apt install pkg-config libgl1-mesa-dev xorg-dev libxkbcommon-dev libwayland-dev wayland-protocols

go build .
go test ./internal/...
go vet ./internal/...
gofmt -l internal assets main.go   # must print nothing
```

Guidelines:

- Keep packages decoupled: providers implement interfaces (`search.Provider`), the UI never reaches into storage directly.
- Every user-facing string goes through `i18n.T(...)` — never hardcode UI text.
- Commands are executed as argv, never through a shell.
- New logic comes with unit tests. **Every logic package must keep at least 80% coverage** — CI enforces it. `internal/ui` is exempt because it is Fyne rendering glue; when you add logic there, extract it into a testable unit (see `terminal.go`, `layout.go`, `hotkeycapture.go`) instead of burying it in widget wiring.
- Commit messages follow [Conventional Commits](https://www.conventionalcommits.org) (`feat:`, `fix:`, `docs:`…), in English.

`third_party/systray` and `third_party/hotkey` are vendored dependencies carrying documented patches (each one marked `laucha patch` in the source) — sync carefully with upstream if you touch them.

## Workflow

laucha uses GitHub Flow — no develop branch:

- `main` is always releasable; releases are tags on `main`.
- Every change lives in a short branch (`feature/...`, `fix/...`, `docs/...`) and lands through a pull request with CI green.
- External contributors: fork, branch, pull request.
- Keep PRs focused — one change per PR reviews faster.

## Reporting issues

Open a GitHub issue with your distribution, desktop environment (X11/Wayland), what you did, what you expected, and what happened. The log output from running `laucha` in a terminal helps a lot.
