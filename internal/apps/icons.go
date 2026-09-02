package apps

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// iconResolver turns desktop-entry Icon values into image files.
//
// Applications name their icon after the active icon theme
// (system-file-manager, computer, folder-open); only some ship it under
// hicolor. The resolver therefore indexes every themed icon once —
// tens of thousands of files in a few tens of milliseconds — and then
// answers each lookup from memory.
type iconResolver struct {
	index map[string]string
}

func newIconResolver() *iconResolver {
	r := &iconResolver{index: map[string]string{}}
	preferred := preferredThemes()
	best := map[string]int{}

	for _, root := range iconRoots() {
		filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext != ".svg" && ext != ".png" {
				return nil // the toolkit cannot draw .xpm and friends
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return nil
			}
			name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
			score := iconScore(rel, ext, preferred)
			if current, seen := best[name]; !seen || score > current {
				best[name] = score
				r.index[name] = path
			}
			return nil
		})
	}
	return r
}

// resolve returns the image file for an Icon value, empty when the
// icon is unknown.
func (r *iconResolver) resolve(name string) string {
	if name == "" {
		return ""
	}
	if filepath.IsAbs(name) {
		if fileExists(name) {
			return name
		}
		return ""
	}
	// A few entries name the icon with its extension.
	switch ext := strings.ToLower(filepath.Ext(name)); ext {
	case ".svg", ".png", ".xpm":
		name = name[:len(name)-len(ext)]
	}
	if path, ok := r.index[name]; ok {
		return path
	}
	for _, ext := range []string{".svg", ".png"} {
		if candidate := filepath.Join("/usr/share/pixmaps", name+ext); fileExists(candidate) {
			return candidate
		}
	}
	return ""
}

// iconRoots lists the directories holding icon themes.
func iconRoots() []string {
	var roots []string
	if home, err := os.UserHomeDir(); err == nil {
		roots = append(roots, filepath.Join(home, ".icons")) // legacy, still searched
	}
	for _, base := range dataDirs() {
		roots = append(roots, filepath.Join(base, "icons"))
	}
	return roots
}

// iconScore ranks a candidate: the themes the user actually sees win,
// then the largest artwork, then a scalable format.
func iconScore(rel, ext string, preferred []string) int {
	segments := strings.Split(rel, string(filepath.Separator))
	theme := ""
	if len(segments) > 0 {
		theme = segments[0]
	}
	rank := 0
	for i, name := range preferred {
		if strings.EqualFold(name, theme) {
			rank = len(preferred) - i
			break
		}
	}
	size := 0
	for _, segment := range segments {
		if segment == "scalable" {
			size = 512
			break
		}
		if width, _, found := strings.Cut(segment, "x"); found {
			if n, err := strconv.Atoi(width); err == nil {
				size = n
				break
			}
		}
	}
	format := 0
	if ext == ".svg" {
		format = 1
	}
	return rank*1_000_000 + size*10 + format
}

// preferredThemes returns the icon themes to favor: the one the
// desktop is configured with, then the freedesktop fallbacks.
func preferredThemes() []string {
	themes := []string{}
	if theme := configuredIconTheme(); theme != "" {
		themes = append(themes, theme)
	}
	return append(themes, "hicolor", "Adwaita")
}

// configuredIconTheme reads gtk-icon-theme-name from the GTK settings
// files, which every GTK desktop keeps in sync with its own setting.
func configuredIconTheme() string {
	var candidates []string
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates,
			filepath.Join(home, ".config", "gtk-4.0", "settings.ini"),
			filepath.Join(home, ".config", "gtk-3.0", "settings.ini"),
			filepath.Join(home, ".gtkrc-2.0"),
		)
	}
	candidates = append(candidates, "/etc/gtk-3.0/settings.ini")

	for _, path := range candidates {
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		theme := scanIconTheme(f)
		f.Close()
		if theme != "" {
			return theme
		}
	}
	return ""
}

func scanIconTheme(f *os.File) string {
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		key, value, found := strings.Cut(scanner.Text(), "=")
		if !found || strings.TrimSpace(key) != "gtk-icon-theme-name" {
			continue
		}
		return strings.Trim(strings.TrimSpace(value), `"`)
	}
	return ""
}
