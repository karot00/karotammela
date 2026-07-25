package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"goth/internal/config"
	"goth/internal/email"
)

// mockMailer records delivered messages and can be configured to fail.
type mockMailer struct {
	sent []email.ContactMessage
	err  error
}

func (m *mockMailer) Send(_ context.Context, msg email.ContactMessage) error {
	if m.err != nil {
		return m.err
	}
	m.sent = append(m.sent, msg)
	return nil
}

func contactTestHandlers(m MailSender) *Handlers {
	return &Handlers{
		cfg: &config.Config{
			ResendAPIKey:     "re_test",
			ContactFromEmail: "info@karotammela.fi",
			ContactToEmail:   "info@karotammela.fi",
		},
		mailer: m,
	}
}

func postContact(t *testing.T, h *Handlers, body, ip string) *httptest.ResponseRecorder {
	t.Helper()
	if ip == "" {
		ip = "203.0.113.1"
	}
	req := httptest.NewRequest(http.MethodPost, "/api/contact", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", ip)
	rec := httptest.NewRecorder()
	h.Contact(rec, req)
	return rec
}

func validContactBody(t *testing.T, mutate map[string]string) string {
	t.Helper()
	payload := map[string]string{
		"name":    "Test Person",
		"email":   "visitor@example.com",
		"message": "Hello, I would like to talk.",
	}
	for k, v := range mutate {
		payload[k] = v
	}
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// TestContactValidationMatrix ports the zod requestSchema boundaries from
// src/app/api/contact/route.ts: every violation is a 400 with the reference
// error body, and trimming happens before limits are measured.
func TestContactValidationMatrix(t *testing.T) {
	mailer := &mockMailer{}
	h := contactTestHandlers(mailer)

	longName := strings.Repeat("a", 81)
	longEmail := strings.Repeat("a", 196) + "@b.co" // 201 chars
	longMessage := strings.Repeat("m", 4001)
	longCompany := strings.Repeat("c", 121)

	cases := []struct {
		name string
		body string
	}{
		{"malformed json", `{"name":`},
		{"empty object", `{}`},
		{"name too short", validContactBody(t, map[string]string{"name": "A"})},
		{"name too long", validContactBody(t, map[string]string{"name": longName})},
		{"name whitespace-only trims to short", validContactBody(t, map[string]string{"name": "  x "})},
		{"email missing", validContactBody(t, map[string]string{"email": ""})},
		{"email not an email", validContactBody(t, map[string]string{"email": "not-an-email"})},
		{"email no TLD", validContactBody(t, map[string]string{"email": "a@b"})},
		{"email single-char TLD", validContactBody(t, map[string]string{"email": "a@b.c"})},
		{"email too long", validContactBody(t, map[string]string{"email": longEmail})},
		{"message too short", validContactBody(t, map[string]string{"message": "short"})},
		{"message too long", validContactBody(t, map[string]string{"message": longMessage})},
		{"message trims below min", validContactBody(t, map[string]string{"message": "   pad   "})},
		{"company too long", validContactBody(t, map[string]string{"company": longCompany})},
		{"honeypot over one char is a schema violation", validContactBody(t, map[string]string{"website": "xy"})},
	}

	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Unique IP per case so the shared rate limiter never interferes.
			rec := postContact(t, h, tc.body, "198.51.100."+strings.Repeat("1", i+1))
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400; body: %s", rec.Code, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), `"error":"Invalid contact payload."`) {
				t.Errorf("body = %s, want reference 400 payload", rec.Body.String())
			}
		})
	}
	if len(mailer.sent) != 0 {
		t.Errorf("mailer received %d messages, want 0", len(mailer.sent))
	}
}

// TestValidContactEmail checks the ported zod email shape directly.
func TestValidContactEmail(t *testing.T) {
	valid := []string{
		"a@b.co",
		"first.last+tag@sub.domain.io",
		"a_b-c@example.com",
		"x@y.fi",
		"a@b.c-d.com",
	}
	invalid := []string{
		"",
		".a@b.co",        // leading dot
		"a..b@c.co",      // consecutive dots
		"a.@c.co",        // trailing dot in local part
		"a@b",            // no TLD
		"a@b.c",          // TLD too short
		"@b.com",         // empty local part
		"a b@c.com",      // space
		"ä@example.com",  // non-ASCII local (zod regex is A-Z/i only)
		"a@ex_ample.com", // underscore not allowed in domain
		"plainaddress",   // no @
		"a@.com",         // empty domain label
		"a@@b.com",       // double @
		"a@b..com",       // consecutive dots in domain
	}
	for _, s := range valid {
		if !validContactEmail(s) {
			t.Errorf("validContactEmail(%q) = false, want true", s)
		}
	}
	for _, s := range invalid {
		if validContactEmail(s) {
			t.Errorf("validContactEmail(%q) = true, want false", s)
		}
	}
}

// TestContactHoneypot: a filled website field short-circuits with a fake
// success and never delivers (mirrors the reference honeypot branch).
func TestContactHoneypot(t *testing.T) {
	mailer := &mockMailer{}
	h := contactTestHandlers(mailer)

	rec := postContact(t, h, validContactBody(t, map[string]string{"website": "x"}), "192.0.2.10")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if strings.TrimSpace(rec.Body.String()) != `{"ok":true}` {
		t.Errorf("body = %s, want {\"ok\":true}", rec.Body.String())
	}
	if len(mailer.sent) != 0 {
		t.Errorf("honeypot triggered %d deliveries, want 0", len(mailer.sent))
	}
}

// TestContactUnconfigured: no mailer (missing config) maps to the reference
// 503 "Contact delivery is not configured."
func TestContactUnconfigured(t *testing.T) {
	h := contactTestHandlers(nil)

	rec := postContact(t, h, validContactBody(t, nil), "192.0.2.20")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"error":"Contact delivery is not configured."`) {
		t.Errorf("body = %s, want reference 503 payload", rec.Body.String())
	}
}

// TestContactSendFailure: provider errors map to the reference 500 payload.
func TestContactSendFailure(t *testing.T) {
	mailer := &mockMailer{err: errors.New("resend: unexpected status 422")}
	h := contactTestHandlers(mailer)

	rec := postContact(t, h, validContactBody(t, nil), "192.0.2.30")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"error":"Failed to send contact email."`) {
		t.Errorf("body = %s, want reference 500 payload", rec.Body.String())
	}
}

// TestContactSuccess verifies the exact Resend message the reference builds:
// from/to from config, replyTo = visitor email, subject prefix, and the
// Name/Email[/Company]/blank-line/message text body.
func TestContactSuccess(t *testing.T) {
	t.Run("with company", func(t *testing.T) {
		mailer := &mockMailer{}
		h := contactTestHandlers(mailer)
		body := validContactBody(t, map[string]string{
			"name":    "  Karo Tammela  ",
			"company": "Staffbite",
			"message": "  Tähän viesti suomeksi.  ",
		})
		rec := postContact(t, h, body, "192.0.2.40")
		if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != `{"ok":true}` {
			t.Fatalf("status = %d body = %s, want 200 {\"ok\":true}", rec.Code, rec.Body.String())
		}
		if len(mailer.sent) != 1 {
			t.Fatalf("deliveries = %d, want 1", len(mailer.sent))
		}
		got := mailer.sent[0]
		want := email.ContactMessage{
			From:    "info@karotammela.fi",
			To:      "info@karotammela.fi",
			ReplyTo: "visitor@example.com",
			Subject: "karotammela.fi contact: Karo Tammela",
			Text:    "Name: Karo Tammela\nEmail: visitor@example.com\nCompany: Staffbite\n\nTähän viesti suomeksi.",
		}
		if got != want {
			t.Errorf("message = %+v\nwant %+v", got, want)
		}
	})

	t.Run("without company", func(t *testing.T) {
		mailer := &mockMailer{}
		h := contactTestHandlers(mailer)
		rec := postContact(t, h, validContactBody(t, nil), "192.0.2.41")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if len(mailer.sent) != 1 {
			t.Fatalf("deliveries = %d, want 1", len(mailer.sent))
		}
		wantText := "Name: Test Person\nEmail: visitor@example.com\n\nHello, I would like to talk."
		if mailer.sent[0].Text != wantText {
			t.Errorf("text = %q\nwant %q", mailer.sent[0].Text, wantText)
		}
	})
}

// TestContactRateLimit ports the reference limit: 8 requests per IP per
// minute, then 429 with Retry-After. Invalid payloads consume quota too
// because the limit is enforced before parsing.
func TestContactRateLimit(t *testing.T) {
	mailer := &mockMailer{}
	h := contactTestHandlers(mailer)
	ip := "192.0.2.99"

	for i := 1; i <= 8; i++ {
		body := validContactBody(t, nil)
		if i%2 == 0 {
			body = `{"name":` // malformed: still consumes quota
		}
		rec := postContact(t, h, body, ip)
		if rec.Code == http.StatusTooManyRequests {
			t.Fatalf("request %d rate-limited early", i)
		}
	}

	rec := postContact(t, h, validContactBody(t, nil), ip)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("9th request status = %d, want 429", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"error":"Rate limit exceeded. Retry shortly."`) {
		t.Errorf("body = %s, want reference 429 payload", rec.Body.String())
	}
	if ra := rec.Header().Get("Retry-After"); ra == "" || ra == "0" {
		t.Errorf("Retry-After = %q, want positive seconds", ra)
	}
}
