package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	blogcontent "goth/content"
	"goth/internal/ai"
	"goth/internal/security"
)

type vipChatRequest struct {
	Messages []ai.Message `json:"messages"`
}

func (h *Handlers) VIPChat(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.VIPEnabled || !h.vipAuthorized(r) {
		h.vipNotFound(w, r)
		return
	}
	h.vipNoStoreHeaders(w)
	if r.Method != http.MethodPost || !h.vipBrowserOriginOK(r) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if h.vipGemini == nil || h.cfg.GoogleAPIKey == "" {
		h.vipSSEError(w, "AI concierge is temporarily unavailable.")
		return
	}
	if h.vipChat == nil {
		h.vipChat = security.NewVIPChatLimiter()
	}
	cookie, _ := r.Cookie(security.VIPCookieName)
	key := vipHash(security.GetClientIP(headerMap(r)) + ":" + cookie.Value)
	if ok, retry := h.vipChat.Allow(key, time.Now()); !ok {
		h.HeaderRateLimitedSSE(w, retry, "The concierge is taking a short break. Please try again later.")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, vipMaxBodyBytes)
	defer r.Body.Close()
	var req vipChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.vipSSEErrorStatus(w, http.StatusBadRequest, "The conversation could not be read.")
		return
	}
	history, err := ai.NormalizeConciergeHistory(req.Messages)
	totalChars := 0
	for _, message := range history {
		totalChars += len(message.Content)
	}
	if err != nil || totalChars > ai.ConciergeMaxInput {
		h.vipSSEErrorStatus(w, http.StatusBadRequest, "Please send a shorter conversation.")
		return
	}
	dossier, err := blogcontent.FS.ReadFile("vip/dossier.md")
	if err != nil {
		h.vipSSEError(w, "The concierge is temporarily unavailable.")
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	writeVIPEvent(w, "meta", map[string]string{"state": "streaming"})
	flusher.Flush()
	start := time.Now()
	_, err = h.vipGemini.StreamWithPrompt(ai.StreamOptions{
		History:         history,
		SystemPrompt:    ai.ConciergeSystemPrompt(string(dossier), h.cfg.VIPContactEmail, h.cfg.VIPContactPhone),
		Temperature:     0.25,
		MaxOutputTokens: 500,
	}, r.Context(), w, flusher)
	if err != nil {
		log.Printf("vip.chat state=error model=%s duration_ms=%d", h.cfg.VIPAIModel, time.Since(start).Milliseconds())
		writeVIPEvent(w, "error", map[string]string{"error": "The concierge could not complete that answer."})
		flusher.Flush()
		return
	}
	log.Printf("vip.chat state=complete model=%s duration_ms=%d", h.cfg.VIPAIModel, time.Since(start).Milliseconds())
	writeVIPEvent(w, "done", map[string]string{"state": "complete"})
	flusher.Flush()
}

func writeVIPEvent(w io.Writer, event string, value any) {
	b, _ := json.Marshal(value)
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, b)
}

func (h *Handlers) vipSSEError(w http.ResponseWriter, message string) {
	h.vipSSEErrorStatus(w, http.StatusServiceUnavailable, message)
}

func (h *Handlers) vipSSEErrorStatus(w http.ResponseWriter, status int, message string) {
	h.vipNoStoreHeaders(w)
	w.Header().Set("Content-Type", "text/event-stream")
	if status >= 400 {
		w.WriteHeader(status)
	}
	writeVIPEvent(w, "error", map[string]string{"error": strings.TrimSpace(message)})
}

func (h *Handlers) HeaderRateLimitedSSE(w http.ResponseWriter, retryMs int64, message string) {
	h.vipNoStoreHeaders(w)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Retry-After", strconv.FormatInt((retryMs+999)/1000, 10))
	w.WriteHeader(http.StatusTooManyRequests)
	writeVIPEvent(w, "error", map[string]string{"error": message, "state": "rate_limited"})
}
