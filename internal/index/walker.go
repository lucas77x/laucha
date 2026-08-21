package index

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/lucas77x/laucha/internal/launcher"
)

// walk scans the roots and returns the files that pass the filter,
// plus every traversed directory (the watcher needs them).
func walk(roots []string, filter *Filter) (files []launcher.Entry, dirs []string) {
	for _, root := range roots {
		filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				if path != root && !filter.EnterDir(path) {
					return filepath.SkipDir
				}
				dirs = append(dirs, path)
				return nil
			}
			if !d.Type().IsRegular() || !filter.IncludeFile(path) {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			files = append(files, launcher.Entry{
				Kind:    launcher.KindFile,
				Name:    d.Name(),
				Path:    path,
				ModTime: info.ModTime(),
			})
			return nil
		})
	}
	return files, dirs
}

// expandRoots resolves "~" and "~/dir" against the user home.
func expandRoots(roots []string) []string {
	home, err := os.UserHomeDir()
	out := make([]string, 0, len(roots))
	for _, root := range roots {
		if err == nil {
			if root == "~" {
				root = home
			} else if strings.HasPrefix(root, "~/") {
				root = filepath.Join(home, root[2:])
			}
		}
		out = append(out, root)
	}
	return out
}
