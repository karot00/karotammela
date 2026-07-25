package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"goth/internal/security"
)

// enforceIPRateLimit applies a per-IP rate limit for the given scope using the
// shared security limiter. When the bucket is exhausted it logs the redacted
// event tag, sets Retry-After (ceil seconds, min 1), writes a JSON 429 with the
// provided message, and returns false; the caller must stop handling the
// request. On success it returns true and writes nothing.
//
// Only the static event tag is logged — never the client IP, query, or any
// user content — so rate-limit logs stay redaction-safe across every endpoint.
func (h *Handlers) enforceIPRateLimit(w http.ResponseWriter, r *http.Request, scope string, limit int, window time.Duration, event, message string) bool {
	ip := security.GetClientIP(headerMap(r))
	allowed, _, retryAfterMs := security.EnforceRateLimit(scope, ip, limit, window)
	if allowed {
		return true
	}
	log.Printf("%s", event)
	w.Header().Set("Retry-After", strconv.FormatInt((retryAfterMs+999)/1000, 10))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	body, _ := json.Marshal(map[string]string{"error": message})
	w.Write(body)
	return false
}
