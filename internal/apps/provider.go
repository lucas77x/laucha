// Package apps discovers installed applications from freedesktop
// .desktop entries and resolves their icons.
package apps

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lucas77x/laucha/internal/launcher"
)

type Provider struct {
	entries []launcher.Entry
}

// NewProvider scans the standard application directories once at
// startup. Live rescanning is on the roadmap.
func NewProvider() *Provider {
	return &Provider{entries: scan()}
}

func (p *Provider) Entries() []launcher.Entry { return p.entries }

func scan() []launcher.Entry {
	var entries []launcher.Entry
	seen := map[string]bool{} // desktop file IDs; earlier dirs take precedence
	for _, dir := range dataDirs() {
		root := filepath.Join(dir, "applications")
		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".desktop") {
				return nil
			}
			id := filepath.Base(path)
			if seen[id] {
				return nil
			}
			seen[id] = true
			if entry, ok := parseDesktopFile(path); ok {
				entries = append(entries, entry)
			}
			return nil
		})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	return entries
}

func dataDirs() []string {
	var dirs []string
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, ".local", "share"))
	}
	xdg := os.Getenv("XDG_DATA_DIRS")
	if xdg == "" {
		xdg = "/usr/local/share:/usr/share"
	}
	for _, d := range strings.Split(xdg, ":") {
		if d != "" {
			dirs = append(dirs, d)
		}
	}
	return dirs
}

// parseDesktopFile extracts the fields laucha needs from the
// [Desktop Entry] section.
func parseDesktopFile(path string) (launcher.Entry, bool) {
	f, err := os.Open(path)
	if err != nil {
		return launcher.Entry{}, false
	}
	defer f.Close()

	fields := map[string]string{}
	inSection := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[") {
			inSection = line == "[Desktop Entry]"
			continue
		}
		if !inSection {
			continue
		}
		if key, value, found := strings.Cut(line, "="); found {
			if _, dup := fields[key]; !dup {
				fields[key] = value
			}
		}
	}

	launchable := fields["Type"] == "Application" &&
		fields["Name"] != "" &&
		fields["NoDisplay"] != "true" &&
		fields["Hidden"] != "true"
	if !launchable {
		return launcher.Entry{}, false
	}

	argv := parseExec(fields["Exec"])
	if len(argv) == 0 {
		return launcher.Entry{}, false
	}

	return launcher.Entry{
		Kind:     launcher.KindApp,
		Name:     fields["Name"],
		Path:     path,
		Exec:     argv,
		Terminal: fields["Terminal"] == "true",
		Icon:     resolveIcon(fields["Icon"]),
	}, true
}

// parseExec splits a desktop-entry Exec value into argv, honoring the
// spec's double-quote and backslash rules and dropping %-field codes,
// so entries are executed directly and never through a shell.
func parseExec(command string) []string {
	var argv []string
	var current strings.Builder
	inQuotes := false
	escaped := false
	flush := func() {
		if current.Len() > 0 {
			argv = append(argv, current.String())
			current.Reset()
		}
	}
	for _, r := range command {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
		case r == '\\':
			escaped = true
		case r == '"':
			inQuotes = !inQuotes
		case r == ' ' && !inQuotes:
			flush()
		default:
			current.WriteRune(r)
		}
	}
	flush()

	kept := argv[:0]
	for _, arg := range argv {
		if strings.HasPrefix(arg, "%") {
			continue
		}
		kept = append(kept, arg)
	}
	return kept
}

// resolveIcon maps an Icon value to an image file, looking through
// the hicolor theme of every XDG data dir — the same set the
// .desktop scan uses — so flatpak and snap exports are covered too.
// Full icon-theme resolution (themes beyond hicolor) is on the
// roadmap.
func resolveIcon(name string) string {
	if name == "" {
		return ""
	}
	if filepath.IsAbs(name) {
		if fileExists(name) {
			return name
		}
		return ""
	}

	sizes := []string{"scalable", "512x512", "256x256", "128x128", "64x64", "48x48"}
	for _, base := range dataDirs() {
		for _, size := range sizes {
			for _, ext := range []string{".svg", ".png"} {
				candidate := filepath.Join(base, "icons", "hicolor", size, "apps", name+ext)
				if fileExists(candidate) {
					return candidate
				}
			}
		}
	}

	candidates := []string{name + ".svg", name + ".png"}
	if ext := filepath.Ext(name); ext == ".svg" || ext == ".png" {
		candidates = []string{name}
	}
	for _, c := range candidates {
		candidate := filepath.Join("/usr/share/pixmaps", c)
		if fileExists(candidate) {
			return candidate
		}
	}
	return ""
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
