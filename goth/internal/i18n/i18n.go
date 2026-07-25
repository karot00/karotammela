package i18n

import (
	"embed"
	"encoding/json"
	"strings"
	"sync"
)

//go:embed messages/*.json
var messagesFS embed.FS

var (
	store     = map[string]map[string]any{}
	storeOnce sync.Once
)

// Locales supported by the app.
var Locales = []string{"en", "fi"}

// DefaultLocale is used when none / invalid is provided.
const DefaultLocale = "fi"

func load() {
	storeOnce.Do(func() {
		for _, loc := range Locales {
			data, err := messagesFS.ReadFile("messages/" + loc + ".json")
			if err != nil {
				continue
			}
			var parsed map[string]any
			if json.Unmarshal(data, &parsed) == nil {
				store[loc] = parsed
			}
		}
	})
}

func lookup(m map[string]any, path string) (string, bool) {
	parts := strings.Split(path, ".")
	var cur any = m
	for _, p := range parts {
		asMap, ok := cur.(map[string]any)
		if !ok {
			return "", false
		}
		var exists bool
		cur, exists = asMap[p]
		if !exists {
			return "", false
		}
	}
	if s, ok := cur.(string); ok {
		return s, true
	}
	return "", false
}

// T returns the translated string for a dotted key in the given locale.
// Falls back to the default locale, then to the key itself. A key that
// resolves to a genuinely empty string (e.g. an intentionally blank
// description) is a valid value and is returned as-is rather than
// triggering the default-locale/key fallback.
func T(locale, key string) string {
	load()
	if m, ok := store[locale]; ok {
		if v, found := lookup(m, key); found {
			return v
		}
	}
	if m, ok := store[DefaultLocale]; ok {
		if v, found := lookup(m, key); found {
			return v
		}
	}
	return key
}

// Exists reports whether a locale is valid.
func Exists(locale string) bool {
	load()
	_, ok := store[locale]
	return ok
}
