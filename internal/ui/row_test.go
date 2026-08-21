package ui

import (
	"path/filepath"
	"testing"
)

func TestFileIconCategories(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"informe.pdf", "file-pdf.svg"},
		{"notas.TXT", "file-text.svg"},
		{"foto.jpeg", "file-image.svg"},
		{"main.go", "file-code.svg"},
		{"algo.raro", "file-generic.svg"},
		{"sin-extension", "file-generic.svg"},
	}
	for _, c := range cases {
		res := fileIcon(c.name)
		if res.Name() != c.want {
			t.Errorf("fileIcon(%q) = %s, want %s", c.name, res.Name(), c.want)
		}
	}
}

func TestFileIconCachesResources(t *testing.T) {
	first := fileIcon("a.pdf")
	second := fileIcon("b.pdf")
	if first != second {
		t.Error("same category returned different resources; cache broken")
	}
}

func TestDisplayDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if got := displayDir(filepath.Join(home, "docs", "x.txt")); got != "~/docs" {
		t.Errorf("displayDir inside home = %q, want ~/docs", got)
	}
	if got := displayDir(filepath.Join(home, "x.txt")); got != "~" {
		t.Errorf("displayDir at home root = %q, want ~", got)
	}
	if got := displayDir("/etc/hosts"); got != "/etc" {
		t.Errorf("displayDir outside home = %q, want /etc", got)
	}
}
