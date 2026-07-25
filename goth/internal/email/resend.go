// Package email provides transactional email delivery via the Resend REST API.
// Direct REST is used instead of the SDK to keep the binary small (plan §6).
package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.resend.com"

// ContactMessage is a single transactional contact email.
type ContactMessage struct {
	From    string
	To      string
	ReplyTo string
	Subject string
	Text    string
}

// ResendSender sends email via the Resend REST API (`POST /emails`).
type ResendSender struct {
	APIKey  string
	BaseURL string // optional override (tests); defaults to the Resend API
	Client  *http.Client
}

type resendEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	ReplyTo string   `json:"reply_to,omitempty"`
	Subject string   `json:"subject"`
	Text    string   `json:"text"`
}

// Send delivers one email. A non-2xx provider response is an error; the
// provider body is deliberately discarded so provider echoes of user input
// never reach logs.
func (s *ResendSender) Send(ctx context.Context, msg ContactMessage) error {
	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	base := s.BaseURL
	if base == "" {
		base = defaultBaseURL
	}

	payload, err := json.Marshal(resendEmailRequest{
		From:    msg.From,
		To:      []string{msg.To},
		ReplyTo: msg.ReplyTo,
		Subject: msg.Subject,
		Text:    msg.Text,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(base, "/")+"/emails", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("resend: unexpected status %d", resp.StatusCode)
	}
	return nil
}
