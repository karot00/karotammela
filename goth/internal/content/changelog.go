package content

import (
	"encoding/json"
	"io/fs"
	"strings"
	"time"

	blogcontent "goth/content"
)

// ChangelogChange is a single entry under a release.
type ChangelogChange struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ChangelogRelease is one versioned release.
type ChangelogRelease struct {
	Version string            `json:"version"`
	Date    string            `json:"date"`
	Title   string            `json:"title"`
	Changes []ChangelogChange `json:"changes"`
}

// Changelog is the parsed changelog payload for a locale.
type Changelog struct {
	Releases []ChangelogRelease `json:"releases"`
}

// GetChangelog loads the localized changelog. Files are authored newest-first,
// so no sorting is applied (mirrors the Next.js reference behavior).
func GetChangelog(locale string) (Changelog, error) {
	if locale != "en" && locale != "fi" {
		locale = "fi"
	}
	path := "changelog/" + locale + ".json"
	raw, err := fs.ReadFile(blogcontent.FS, path)
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			return Changelog{}, nil
		}
		return Changelog{}, err
	}
	var c Changelog
	if err := json.Unmarshal(raw, &c); err != nil {
		return Changelog{}, err
	}
	return c, nil
}

// FormatDate renders a YYYY-MM-DD string with medium date style for the locale.
// On unparseable input it returns the raw string (parity with the Next.js UI).
func FormatDate(locale, date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return localizedDate(locale, t)
}

// localizedDate formats a time with medium date style per locale.
func localizedDate(locale string, t time.Time) string {
	if locale == "fi" {
		return t.Format("02.01.2006")
	}
	return t.Format("Jan 2, 2006")
}

// FormatDateTime renders a timestamp with medium date + short time for the
// dashboard overview "latest unlock" field.
func FormatDateTime(locale string, t time.Time) string {
	if locale == "fi" {
		return t.Format("02.01.2006 15:04")
	}
	return t.Format("Jan 2, 2006 15:04")
}
