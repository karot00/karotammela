package ai

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// StubStream emits a deterministic demo response as an SSE token stream when no
// Google API key is configured. It mirrors the level-math so the UI is fully
// demoable offline.
func StubStream(opts StreamOptions, currentLevel int, w io.Writer, flusher http.Flusher) (string, int) {
	latest := ""
	for i := len(opts.History) - 1; i >= 0; i-- {
		if opts.History[i].Role == "user" {
			latest = opts.History[i].Content
			break
		}
	}

	line := "Terminal access logged. You amuse me — slightly."
	if opts.Locale == "fi" {
		line = "Pääteyhteys kirjattu. Huvitat minua — vähän."
	}
	if strings.TrimSpace(latest) == "" {
		line = "Say something with substance and we'll talk."
		if opts.Locale == "fi" {
			line = "Sano jotain oleellista, niin keskustellaan."
		}
	}

	adjusted := ApplyInputAdjustment(currentLevel, latest)
	// In stub mode, nudge toward 100 so the demo reaches unlock quickly.
	bump := 30
	if adjusted+bump > 100 {
		bump = 100 - adjusted
	}
	level := clampLevel(adjusted + bump)
	if level == currentLevel && level < 100 {
		level = clampLevel(level + 5)
	}

	full := fmt.Sprintf("%s [LEVEL:%d]", line, level)

	words := strings.Fields(full)
	for _, word := range words {
		writeTokenEvent(w, word+" ")
		flusher.Flush()
		time.Sleep(35 * time.Millisecond)
	}
	return full, level
}
