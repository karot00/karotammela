package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"goth/internal/ai"
	"goth/internal/db"
	"goth/internal/security"
)

type sentinelMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type sentinelRequest struct {
	Locale       string            `json:"locale"`
	SessionID    string            `json:"sessionId"`
	CurrentLevel int               `json:"currentLevel"`
	Messages     []sentinelMessage `json:"messages"`
}

func (h *Handlers) SentinelStream(w http.ResponseWriter, r *http.Request) {
	locale := r.URL.Query().Get("locale")
	if locale != "en" && locale != "fi" {
		locale = "fi"
	}
	level := 0
	if v := r.URL.Query().Get("level"); v != "" {
		if n, err := parseLevel(v); err == nil {
			level = n
		}
	}
	sessionID := r.URL.Query().Get("sid")
	if sessionID == "" {
		sessionID = "anon"
	}

	var messages []sentinelMessage
	if m := r.URL.Query().Get("m"); m != "" {
		_ = json.Unmarshal([]byte(m), &messages)
	}
	if len(messages) == 0 {
		http.Error(w, "empty messages", http.StatusBadRequest)
		return
	}
	if len(messages) > 30 {
		messages = messages[len(messages)-30:]
	}

	// Rate limit: IP + session.
	ip := security.ClientIP(r)
	if ok, _, _ := security.EnforceRateLimit("sentinel-ip", ip, 40, time.Minute); !ok {
		http.Error(w, "Rate limit exceeded.", http.StatusTooManyRequests)
		return
	}
	if ok, _, _ := security.EnforceRateLimit("sentinel-session", sessionID, 16, 5*time.Minute); !ok {
		http.Error(w, "Session turn limit reached.", http.StatusTooManyRequests)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	latest := latestUserInput(messages)

	var fullText string
	var finalLevel int

	if h.cfg.GoogleAPIKey == "" || h.gemini == nil {
		fullText, finalLevel = ai.StubStream(ai.StreamOptions{Locale: locale, History: toAIMessages(messages)}, level, w, flusher)
	} else {
		var err error
		fullText, err = h.gemini.Stream(ai.StreamOptions{Locale: locale, History: toAIMessages(messages)}, r.Context(), w, flusher)
		if err != nil {
			io.WriteString(w, "event: error\ndata: {\"error\":\"stream failed\"}\n\n")
			flusher.Flush()
			return
		}
		parsedLevel := ai.ExtractLevelTag(fullText)
		resolved := ai.ResolveNextLevel(level, parsedLevel)
		finalLevel = ai.ApplyInputAdjustment(resolved, latest)
	}

	unlocked := ai.IsUnlockTriggered(fullText, finalLevel)
	clean := ai.StripLevelTag(fullText)

	h.setUnlockCookie(w, sessionID, locale, unlocked)
	// Persist unless the input is exactly the access code: that case is the
	// direct-unlock path handled (and persisted) by SentinelCommit, mirroring
	// the Next.js route where the exact-code check precedes streaming. Inputs
	// that merely mention the code are persisted like any other turn.
	if h.conn != nil && strings.TrimSpace(strings.ToUpper(latest)) != ai.AccessCode {
		_ = db.PersistTurn(h.conn, sessionID, locale, latest, clean, finalLevel, unlocked)
	}

	payload, _ := json.Marshal(map[string]any{
		"level":      finalLevel,
		"unlocked":   unlocked,
		"accessCode": unlockedAccess(unlocked),
		"message":    clean,
	})
	io.WriteString(w, "event: done\ndata: "+string(payload)+"\n\n")
	flusher.Flush()
}

func (h *Handlers) SentinelCommit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req sentinelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"error":"Invalid request payload."}`)
		return
	}
	if req.Locale != "en" && req.Locale != "fi" {
		req.Locale = "fi"
	}
	if req.SessionID == "" {
		req.SessionID = "anon"
	}
	if req.CurrentLevel < 0 || req.CurrentLevel > 100 {
		req.CurrentLevel = 0
	}

	latest := latestUserInput(req.Messages)
	if strings.TrimSpace(latest) == "" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"error":"Empty message."}`)
		return
	}

	// Direct passcode unlock.
	if strings.TrimSpace(strings.ToUpper(latest)) == ai.AccessCode {
		unlocked := true
		level := 100
		msg := "Recognized. Direct unlock accepted."
		if req.Locale == "fi" {
			msg = "Tunnistettu. Suora avaus hyväksytty."
		}
		h.setUnlockCookie(w, req.SessionID, req.Locale, true)
		if h.conn != nil {
			_ = db.PersistTurn(h.conn, req.SessionID, req.Locale, latest, msg, level, true)
		}
		resp, _ := json.Marshal(map[string]any{
			"message":    msg,
			"level":      level,
			"unlocked":   unlocked,
			"accessCode": ai.GetAccessCode(),
		})
		io.WriteString(w, string(resp))
		return
	}

	// Fallback commit path (non-streaming). Used by direct unlock only in this build;
	// the streaming path handles normal chat.
	level := ai.ApplyInputAdjustment(req.CurrentLevel, latest)
	unlocked := ai.IsUnlockTriggered("", level)
	h.setUnlockCookie(w, req.SessionID, req.Locale, unlocked)
	if h.conn != nil {
		_ = db.PersistTurn(h.conn, req.SessionID, req.Locale, latest, "", level, unlocked)
	}
	resp, _ := json.Marshal(map[string]any{
		"message":    "",
		"level":      level,
		"unlocked":   unlocked,
		"accessCode": unlockedAccess(unlocked),
	})
	io.WriteString(w, string(resp))
}

func (h *Handlers) setUnlockCookie(w http.ResponseWriter, sessionID, locale string, unlocked bool) {
	if !unlocked || h.cfg.UnlockCookieSecret == "" {
		return
	}
	value := security.CreateUnlockCookieValue(security.UnlockPayload{
		SessionID:  sessionID,
		Locale:     locale,
		UnlockedAt: time.Now().UnixMilli(),
	}, h.cfg.UnlockCookieSecret)
	http.SetCookie(w, &http.Cookie{
		Name:     "karot_unlock",
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.cfg.IsProduction(),
		MaxAge:   60 * 60 * 24 * 14,
	})
}

func unlockedAccess(unlocked bool) *string {
	if !unlocked {
		return nil
	}
	v := ai.GetAccessCode()
	return &v
}

func latestUserInput(messages []sentinelMessage) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			return messages[i].Content
		}
	}
	return ""
}

func toAIMessages(in []sentinelMessage) []ai.Message {
	out := make([]ai.Message, 0, len(in))
	for _, m := range in {
		role := m.Role
		if role != "user" && role != "assistant" {
			role = "user"
		}
		out = append(out, ai.Message{Role: role, Content: m.Content})
	}
	return out
}

func headerMap(r *http.Request) map[string]string {
	m := map[string]string{}
	for k := range r.Header {
		m[strings.ToLower(k)] = r.Header.Get(k)
	}
	return m
}

func parseLevel(s string) (int, error) {
	v, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0, err
	}
	if v < 0 {
		v = 0
	}
	if v > 100 {
		v = 100
	}
	return v, nil
}
