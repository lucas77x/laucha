package index

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lucas77x/laucha/internal/config"
)

// Filter decides which paths enter the index, compiled once from the
// user configuration.
type Filter struct {
	includeOnly bool
	extensions  map[string]bool // lowercase, with leading dot
	names       map[string]bool // exact base names
	patterns    []*regexp.Regexp
}

func NewFilter(cfg config.Filter) *Filter {
	f := &Filter{
		includeOnly: cfg.Mode == "include-only",
		extensions:  map[string]bool{},
		names:       map[string]bool{},
	}
	for _, ext := range cfg.Extensions {
		ext = strings.ToLower(ext)
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		f.extensions[ext] = true
	}
	for _, name := range cfg.Names {
		f.names[name] = true
	}
	for _, pattern := range cfg.Patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue // invalid user regex: skip it rather than crash
		}
		f.patterns = append(f.patterns, re)
	}
	return f
}

// IncludeFile reports whether a file belongs in the index.
func (f *Filter) IncludeFile(path string) bool {
	if f.includeOnly {
		return f.matches(path)
	}
	return !f.matches(path)
}

// EnterDir reports whether the walker should descend into a
// directory. In include-only mode every directory is traversed,
// since a match may live anywhere below it.
func (f *Filter) EnterDir(path string) bool {
	if f.includeOnly {
		return true
	}
	return !f.matches(path)
}

// IncludeDir reports whether a directory itself belongs in the index,
// as opposed to merely being traversed. In exclude mode this mirrors
// EnterDir: a directory is indexed whenever the walker would descend
// into it. In include-only mode a directory is indexed only when its
// base name or path is explicitly matched by names or patterns;
// extensions never apply to directories.
func (f *Filter) IncludeDir(path string) bool {
	if !f.includeOnly {
		return !f.matches(path)
	}
	if f.names[filepath.Base(path)] {
		return true
	}
	for _, re := range f.patterns {
		if re.MatchString(path) {
			return true
		}
	}
	return false
}

// matches reports whether path hits any of the three matcher lists.
// Names match every path segment, so a file inside an excluded
// directory is excluded even when the event skips the walker's
// pruning.
func (f *Filter) matches(path string) bool {
	if f.extensions[strings.ToLower(filepath.Ext(path))] {
		return true
	}
	if len(f.names) > 0 {
		for _, segment := range strings.Split(path, "/") {
			if f.names[segment] {
				return true
			}
		}
	}
	for _, re := range f.patterns {
		if re.MatchString(path) {
			return true
		}
	}
	return false
}
