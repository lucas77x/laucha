package ui

import (
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"

	"github.com/lucas77x/laucha/assets"
)

// extCategory groups extensions into the bundled file-icon set, so
// results look consistent on every desktop theme.
var extCategory = map[string]string{
	".txt": "text", ".md": "text", ".log": "text", ".rtf": "text",
	".pdf": "pdf",
	".doc": "doc", ".docx": "doc", ".odt": "doc",
	".xls": "sheet", ".xlsx": "sheet", ".ods": "sheet", ".csv": "sheet",
	".png": "image", ".jpg": "image", ".jpeg": "image", ".gif": "image",
	".svg": "image", ".webp": "image", ".bmp": "image", ".ico": "image",
	".mp3": "audio", ".wav": "audio", ".flac": "audio", ".ogg": "audio", ".m4a": "audio",
	".mp4": "video", ".mkv": "video", ".avi": "video", ".mov": "video", ".webm": "video",
	".zip": "archive", ".tar": "archive", ".gz": "archive", ".bz2": "archive",
	".xz": "archive", ".7z": "archive", ".rar": "archive", ".deb": "archive",
	".go": "code", ".js": "code", ".ts": "code", ".py": "code", ".rs": "code",
	".c": "code", ".cpp": "code", ".h": "code", ".java": "code", ".sh": "code",
	".html": "code", ".css": "code", ".json": "code", ".yaml": "code",
	".yml": "code", ".toml": "code", ".sql": "code",
}

// iconCache is only touched from the UI event loop.
var iconCache = map[string]fyne.Resource{}

// fileIcon picks the bundled icon for a file name by extension.
func fileIcon(name string) fyne.Resource {
	category := extCategory[strings.ToLower(filepath.Ext(name))]
	if category == "" {
		category = "generic"
	}
	key := "file-" + category
	if res, ok := iconCache[key]; ok {
		return res
	}
	data, ok := assets.FileIcon(key)
	if !ok {
		data, _ = assets.FileIcon("file-generic")
	}
	res := fyne.NewStaticResource(key+".svg", data)
	iconCache[key] = res
	return res
}

// displayDir shows where a file lives, contracted against the home
// directory (~/desarrollo/foo).
func displayDir(path string) string {
	dir := filepath.Dir(path)
	home, err := os.UserHomeDir()
	if err != nil {
		return dir
	}
	if dir == home {
		return "~"
	}
	if rel, err := filepath.Rel(home, dir); err == nil && !strings.HasPrefix(rel, "..") {
		return "~/" + rel
	}
	return dir
}
