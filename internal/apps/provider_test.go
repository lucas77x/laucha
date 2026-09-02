package apps

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeDesktop(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestScanParsesLaunchableEntries(t *testing.T) {
	data := t.TempDir()
	apps := filepath.Join(data, "applications")
	writeDesktop(t, apps, "editor.desktop", `[Desktop Entry]
Type=Application
Name=Cool Editor
Exec=cooledit %F --flag
Icon=/nonexistent/icon.png
`)
	writeDesktop(t, apps, "hidden.desktop", `[Desktop Entry]
Type=Application
Name=Hidden
Exec=hidden
NoDisplay=true
`)
	writeDesktop(t, apps, "terminal-app.desktop", `[Desktop Entry]
Type=Application
Name=Top
Exec=htop
Terminal=true
`)
	writeDesktop(t, apps, "service.desktop", `[Desktop Entry]
Type=Service
Name=Not An App
Exec=svc
`)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_DIRS", data)

	entries := NewProvider().Entries()

	byName := map[string]int{}
	for i, e := range entries {
		byName[e.Name] = i
	}
	if _, ok := byName["Cool Editor"]; !ok {
		t.Fatalf("Cool Editor missing from %d entries", len(entries))
	}
	editor := entries[byName["Cool Editor"]]
	if len(editor.Exec) != 2 || editor.Exec[0] != "cooledit" || editor.Exec[1] != "--flag" {
		t.Errorf("Exec = %v, want [cooledit --flag] with %%F stripped", editor.Exec)
	}
	if editor.Terminal {
		t.Error("Cool Editor marked Terminal")
	}
	if _, ok := byName["Hidden"]; ok {
		t.Error("NoDisplay entry leaked into results")
	}
	if _, ok := byName["Not An App"]; ok {
		t.Error("non-Application entry leaked into results")
	}
	top, ok := byName["Top"]
	if !ok {
		t.Fatal("Terminal=true app missing: terminal apps must be listed")
	}
	if !entries[top].Terminal {
		t.Error("Top not marked as terminal app")
	}
}

func TestScanDeduplicatesById(t *testing.T) {
	dataA := t.TempDir()
	dataB := t.TempDir()
	writeDesktop(t, filepath.Join(dataA, "applications"), "app.desktop", `[Desktop Entry]
Type=Application
Name=First Wins
Exec=first
`)
	writeDesktop(t, filepath.Join(dataB, "applications"), "app.desktop", `[Desktop Entry]
Type=Application
Name=Second Loses
Exec=second
`)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_DIRS", dataA+":"+dataB)

	entries := NewProvider().Entries()

	if len(entries) != 1 || entries[0].Name != "First Wins" {
		t.Errorf("entries = %v, want only First Wins", entries)
	}
}

func TestParseExecQuoting(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{`prog --flag`, []string{"prog", "--flag"}},
		{`prog %U %f`, []string{"prog"}},
		{`"my prog" arg`, []string{"my prog", "arg"}},
		{`env VAR=1 prog`, []string{"env", "VAR=1", "prog"}},
		{`prog \"quoted`, []string{"prog", `"quoted`}},
		{``, nil},
	}
	for _, c := range cases {
		got := parseExec(c.in)
		if len(got) != len(c.want) {
			t.Errorf("parseExec(%q) = %v, want %v", c.in, got, c.want)
			continue
		}
		for i := range c.want {
			if got[i] != c.want[i] {
				t.Errorf("parseExec(%q)[%d] = %q, want %q", c.in, i, got[i], c.want[i])
			}
		}
	}
}

func writeIconFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("icon"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestIconResolverAbsolutePaths(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_DIRS", t.TempDir())
	r := newIconResolver()

	if got := r.resolve(""); got != "" {
		t.Errorf("resolve(empty) = %q, want empty", got)
	}
	if got := r.resolve("/definitely/missing/icon.png"); got != "" {
		t.Errorf("resolve(missing absolute) = %q, want empty", got)
	}
	abs := filepath.Join(t.TempDir(), "icon.png")
	writeIconFile(t, abs)
	if got := r.resolve(abs); got != abs {
		t.Errorf("resolve(existing absolute) = %q, want %q", got, abs)
	}
}

func TestIconResolverFindsIconsOutsideHicolor(t *testing.T) {
	home := t.TempDir()
	data := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_DIRS", data)
	writeIconFile(t, filepath.Join(home, ".config", "gtk-3.0", "settings.ini"))
	if err := os.WriteFile(filepath.Join(home, ".config", "gtk-3.0", "settings.ini"),
		[]byte("[Settings]\ngtk-icon-theme-name=MyTheme\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// The same name in the configured theme and in hicolor.
	writeIconFile(t, filepath.Join(data, "icons", "MyTheme", "48x48", "apps", "editor.png"))
	writeIconFile(t, filepath.Join(data, "icons", "hicolor", "48x48", "apps", "editor.png"))
	// Names desktop entries borrow from the theme live outside apps/.
	writeIconFile(t, filepath.Join(data, "icons", "OtherTheme", "scalable", "places", "folder-open.svg"))
	// Dotted names must survive the extension trimming.
	writeIconFile(t, filepath.Join(data, "icons", "hicolor", "scalable", "apps", "com.example.App.svg"))

	r := newIconResolver()

	if got := r.resolve("editor"); !strings.Contains(got, "MyTheme") {
		t.Errorf("resolve(editor) = %q, want the configured theme to win", got)
	}
	if got := r.resolve("folder-open"); !strings.HasSuffix(got, "folder-open.svg") {
		t.Errorf("resolve(folder-open) = %q, want the themed places icon", got)
	}
	if got := r.resolve("editor.png"); !strings.HasSuffix(got, "editor.png") {
		t.Errorf("resolve(editor.png) = %q, want the extension to be trimmed", got)
	}
	if got := r.resolve("com.example.App"); !strings.HasSuffix(got, "com.example.App.svg") {
		t.Errorf("resolve(com.example.App) = %q, want the dotted name intact", got)
	}
	if got := r.resolve("nothing-like-this"); got != "" {
		t.Errorf("resolve(unknown) = %q, want empty", got)
	}
}

func TestIconScorePrefersThemeThenSize(t *testing.T) {
	preferred := []string{"MyTheme", "hicolor"}
	themed := iconScore(filepath.Join("MyTheme", "48x48", "apps", "a.png"), ".png", preferred)
	fallback := iconScore(filepath.Join("hicolor", "512x512", "apps", "a.png"), ".png", preferred)
	if themed <= fallback {
		t.Error("the configured theme must outrank a bigger icon from another theme")
	}
	small := iconScore(filepath.Join("hicolor", "48x48", "apps", "a.png"), ".png", preferred)
	big := iconScore(filepath.Join("hicolor", "scalable", "apps", "a.svg"), ".svg", preferred)
	if big <= small {
		t.Error("within a theme, scalable artwork must outrank a small bitmap")
	}
}

func TestLiveRescanFollowsInstallAndRemove(t *testing.T) {
	data := t.TempDir()
	apps := filepath.Join(data, "applications")
	if err := os.MkdirAll(apps, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_DIRS", data)

	p := NewProvider()
	defer p.Close()
	if len(p.Entries()) != 0 {
		t.Fatalf("expected empty start, got %v", p.Entries())
	}

	writeDesktop(t, apps, "late.desktop", "[Desktop Entry]\nType=Application\nName=Late App\nExec=late\n")
	waitFor(t, "installed app indexed", func() bool {
		entries := p.Entries()
		return len(entries) == 1 && entries[0].Name == "Late App"
	})

	if err := os.Remove(filepath.Join(apps, "late.desktop")); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "removed app dropped", func() bool {
		return len(p.Entries()) == 0
	})
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}
