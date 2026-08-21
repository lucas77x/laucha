package apps

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func TestResolveIcon(t *testing.T) {
	if got := resolveIcon(""); got != "" {
		t.Errorf("resolveIcon(empty) = %q, want empty", got)
	}
	if got := resolveIcon("/definitely/missing/icon.png"); got != "" {
		t.Errorf("resolveIcon(missing abs) = %q, want empty", got)
	}

	dir := t.TempDir()
	abs := filepath.Join(dir, "icon.png")
	if err := os.WriteFile(abs, []byte("png"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := resolveIcon(abs); got != abs {
		t.Errorf("resolveIcon(existing abs) = %q, want %q", got, abs)
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	themed := filepath.Join(home, ".local", "share", "icons", "hicolor", "48x48", "apps")
	if err := os.MkdirAll(themed, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(themed, "myapp.png"), []byte("png"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := resolveIcon("myapp"); !strings.HasSuffix(got, "myapp.png") {
		t.Errorf("resolveIcon(myapp) = %q, want themed png", got)
	}
}
