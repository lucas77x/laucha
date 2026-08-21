// Package assets embeds the images shipped inside the binary.
package assets

import "embed"

//go:embed icon.svg
var IconSVG []byte

// TraySVG is a simplified high-contrast variant of the app icon that
// stays readable at tiny system-tray sizes.
//
//go:embed tray.svg
var TraySVG []byte

// TrayPNG is the pre-rasterized tray icon; some tray hosts only
// handle raster pixmaps reliably.
//
//go:embed tray.png
var TrayPNG []byte

//go:embed icons
var icons embed.FS

// FileIcon returns the bundled SVG for an icon name such as
// "file-pdf"; ok is false when no such icon exists.
func FileIcon(name string) (data []byte, ok bool) {
	data, err := icons.ReadFile("icons/" + name + ".svg")
	return data, err == nil
}
