package assets

import "embed"

// MediaFS embeds shared media (profile, OG, postcard templates) served from
// /media so the Go binary is self-contained and URL-parity matches Next.js.
//
//go:embed media
var MediaFS embed.FS
