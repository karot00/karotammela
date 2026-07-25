package blogcontent

import "embed"

// FS embeds the blog markdown tree so the binary is self-contained
// (mirrors how web/static and i18n/messages are embedded elsewhere).
//
//go:embed blog
//go:embed changelog
var FS embed.FS
