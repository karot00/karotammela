package blogcontent

import "embed"

// FS embeds the blog markdown tree so the binary is self-contained
// (mirrors how web/static and i18n/messages are embedded elsewhere).
// The vip tree holds the curated MeetingPackage application content
// (plan §8.5); it ships embedded but is only rendered behind the VIP
// access flow.
//
//go:embed blog
//go:embed changelog
//go:embed vip
var FS embed.FS
