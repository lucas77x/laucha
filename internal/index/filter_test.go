package index

import (
	"testing"

	"github.com/lucas77x/laucha/internal/config"
)

func TestExcludeMode(t *testing.T) {
	f := NewFilter(config.Filter{
		Mode:       "exclude",
		Extensions: []string{".log"},
		Names:      []string{"node_modules"},
		Patterns:   []string{`(^|/)\.[^/]+`},
	})

	cases := []struct {
		path string
		want bool
	}{
		{"/home/u/notes.txt", true},
		{"/home/u/app.log", false},
		{"/home/u/.secret", false},
		{"/home/u/.config/x.txt", false},
	}
	for _, c := range cases {
		if got := f.IncludeFile(c.path); got != c.want {
			t.Errorf("IncludeFile(%q) = %v, want %v", c.path, got, c.want)
		}
	}

	if f.EnterDir("/home/u/node_modules") {
		t.Error("EnterDir(node_modules) = true, want false")
	}
	if !f.EnterDir("/home/u/projects") {
		t.Error("EnterDir(projects) = false, want true")
	}
}

func TestIncludeDirExcludeMode(t *testing.T) {
	f := NewFilter(config.Filter{
		Mode:  "exclude",
		Names: []string{"node_modules"},
	})

	if f.IncludeDir("/home/u/node_modules") {
		t.Error("IncludeDir(node_modules) = true, want false")
	}
	if !f.IncludeDir("/home/u/projects") {
		t.Error("IncludeDir(projects) = false, want true")
	}
}

func TestIncludeDirIncludeOnlyMode(t *testing.T) {
	f := NewFilter(config.Filter{
		Mode:       "include-only",
		Extensions: []string{".pdf"},
		Names:      []string{"docs"},
		Patterns:   []string{`^/home/u/photos`},
	})

	if !f.IncludeDir("/home/u/work/docs") {
		t.Error("IncludeDir(docs) = false, want true: matched by name")
	}
	if !f.IncludeDir("/home/u/photos") {
		t.Error("IncludeDir(photos) = false, want true: matched by pattern")
	}
	if f.IncludeDir("/home/u/archive.pdf") {
		t.Error("IncludeDir(archive.pdf) = true, want false: extensions must not apply to directories")
	}
	if f.IncludeDir("/home/u/other") {
		t.Error("IncludeDir(other) = true, want false: unmatched directory")
	}
}

func TestDefaultFilterExcludesCaches(t *testing.T) {
	f := NewFilter(config.Default().Filter)

	cases := []struct {
		path string
		want bool
	}{
		{"/home/u/docs/informe.pdf", true},
		{"/home/u/go/pkg/mod/foo/bar.go", false},
		{"/home/u/snap/code/253/data.txt", false},
		{"/home/u/proyecto/__pycache__/m.pyc", false},
	}
	for _, c := range cases {
		if got := f.IncludeFile(c.path); got != c.want {
			t.Errorf("IncludeFile(%q) = %v, want %v", c.path, got, c.want)
		}
	}
	if f.EnterDir("/home/u/go/pkg") || f.EnterDir("/home/u/snap") {
		t.Error("default filter must prune go/pkg and snap directories")
	}
}

func TestIncludeOnlyMode(t *testing.T) {
	f := NewFilter(config.Filter{
		Mode:       "include-only",
		Extensions: []string{"pdf"}, // no leading dot: must be normalized
	})

	if !f.IncludeFile("/home/u/doc.PDF") {
		t.Error("IncludeFile(doc.PDF) = false, want true")
	}
	if f.IncludeFile("/home/u/doc.txt") {
		t.Error("IncludeFile(doc.txt) = true, want false")
	}
	if !f.EnterDir("/home/u/.hidden") {
		t.Error("EnterDir in include-only mode must always be true")
	}
}

func TestInvalidPatternIsSkipped(t *testing.T) {
	f := NewFilter(config.Filter{Mode: "exclude", Patterns: []string{"(("}})

	if !f.IncludeFile("/home/u/anything.txt") {
		t.Error("an invalid pattern must not exclude everything")
	}
}
