package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"goth/internal/email"
)

// contactRequest mirrors the zod requestSchema in src/app/api/contact/route.ts.
type contactRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
	Company string `json:"company"`
	Website string `json:"website"` // honeypot
}

// Email shape ported from zod's default `.email()` regex: no leading dot, no
// consecutive dots, local part of A-Z0-9_'+-., domain labels starting
// alphanumerically, and a letters-only TLD of 2+ characters. RE2 has no
// lookahead, so the dot rules are checked imperatively in validContactEmail.
var (
	contactEmailLocal  = regexp.MustCompile(`^[A-Za-z0-9_'+\-.]*[A-Za-z0-9_+-]$`)
	contactEmailDomain = regexp.MustCompile(`^([A-Za-z0-9][A-Za-z0-9\-]*\.)+[A-Za-z]{2,}$`)
)

func validContactEmail(s string) bool {
	if s == "" || strings.HasPrefix(s, ".") || strings.Contains(s, "..") {
		return false
	}
	at := strings.LastIndex(s, "@")
	if at < 0 {
		return false
	}
	return contactEmailLocal.MatchString(s[:at]) && contactEmailDomain.MatchString(s[at+1:])
}

// valid reports whether the payload satisfies the reference schema
// (name 2-80, email ≤200 + valid, message 10-4000, company ≤120,
// website ≤1 — all after trimming). Rune counts are used so multibyte
// names/messages measure like JS string lengths for parity.
func (r *contactRequest) valid() bool {
	if n := utf8.RuneCountInString(r.Name); n < 2 || n > 80 {
		return false
	}
	if utf8.RuneCountInString(r.Email) > 200 || !validContactEmail(r.Email) {
		return false
	}
	if n := utf8.RuneCountInString(r.Message); n < 10 || n > 4000 {
		return false
	}
	if utf8.RuneCountInString(r.Company) > 120 {
		return false
	}
	if utf8.RuneCountInString(r.Website) > 1 {
		return false
	}
	return true
}

func writeContactJSON(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	io.WriteString(w, body)
}

func writeContactError(w http.ResponseWriter, status int, msg string) {
	payload, _ := json.Marshal(map[string]string{"error": msg})
	writeContactJSON(w, status, string(payload))
}

// Contact handles POST /api/contact. Port of src/app/api/contact/route.ts:
// per-IP rate limit, schema validation, honeypot short-circuit, config gate,
// then Resend REST delivery. Telemetry events from the reference are emitted
// as redacted log lines (no user content, no addresses).
func (h *Handlers) Contact(w http.ResponseWriter, r *http.Request) {
	if !h.enforceIPRateLimit(w, r, "contact-ip", 8, time.Minute, "contact.rate_limited", "Rate limit exceeded. Retry shortly.") {
		return
	}

	var req contactRequest
	// Bounded body: the schema caps at ~4.3 KB of content; 64 KB is generous.
	// Malformed JSON maps to the same 400 as schema violations (the reference
	// surfaces it as a 500 via request.json(); 400 is the documented contract).
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64*1024)).Decode(&req); err != nil {
		writeContactError(w, http.StatusBadRequest, "Invalid contact payload.")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	req.Message = strings.TrimSpace(req.Message)
	req.Company = strings.TrimSpace(req.Company)
	req.Website = strings.TrimSpace(req.Website)
	if !req.valid() {
		writeContactError(w, http.StatusBadRequest, "Invalid contact payload.")
		return
	}

	if req.Website != "" {
		log.Printf("contact.honeypot_triggered")
		writeContactJSON(w, http.StatusOK, `{"ok":true}`)
		return
	}

	if h.mailer == nil {
		writeContactError(w, http.StatusServiceUnavailable, "Contact delivery is not configured.")
		return
	}

	lines := []string{"Name: " + req.Name, "Email: " + req.Email}
	if req.Company != "" {
		lines = append(lines, "Company: "+req.Company)
	}
	lines = append(lines, "", req.Message)

	err := h.mailer.Send(r.Context(), email.ContactMessage{
		From:    h.cfg.ContactFromEmail,
		To:      h.cfg.ContactToEmail,
		ReplyTo: req.Email,
		Subject: "karotammela.fi contact: " + req.Name,
		Text:    strings.Join(lines, "\n"),
	})
	if err != nil {
		// mailer errors carry at most a provider status code, never content.
		log.Printf("contact.send_failed: %v", err)
		writeContactError(w, http.StatusInternalServerError, "Failed to send contact email.")
		return
	}

	log.Printf("contact.submitted hasCompany=%v", req.Company != "")
	writeContactJSON(w, http.StatusOK, `{"ok":true}`)
}
