// Package assets embeds the images shipped inside the binary.
package assets

import _ "embed"

//go:embed icon.svg
var IconSVG []byte
