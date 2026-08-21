// Package assets embeds the images shipped inside the binary.
package assets

import "embed"

//go:embed icon.svg
var IconSVG []byte

//go:embed icons
var icons embed.FS

// FileIcon returns the bundled SVG for an icon name such as
// "file-pdf"; ok is false when no such icon exists.
func FileIcon(name string) (data []byte, ok bool) {
	data, err := icons.ReadFile("icons/" + name + ".svg")
	return data, err == nil
}
