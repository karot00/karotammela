package webstatic

import "embed"

//go:embed static
var StaticFS embed.FS
