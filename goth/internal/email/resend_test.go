package email

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestResendSendSuccess verifies the REST contract: POST /emails with a Bearer
// token and the exact JSON shape the Resend API expects.
func TestResendSendSuccess(t *testing.T) {
	var gotAuth, gotCT, gotMethod, gotPath string
	var gotBody resendEmailRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotCT = r.Header.Get("Content-Type")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"email_123"}`))
	}))
	defer srv.Close()

	s := &ResendSender{APIKey: "re_test_key", BaseURL: srv.URL + "/"} // trailing slash must be tolerated
	err := s.Send(context.Background(), ContactMessage{
		From:    "info@karotammela.fi",
		To:      "info@karotammela.fi",
		ReplyTo: "visitor@example.com",
		Subject: "karotammela.fi contact: Test Person",
		Text:    "Name: Test Person\nEmail: visitor@example.com\n\nHello there!",
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	if gotMethod != http.MethodPost || gotPath != "/emails" {
		t.Errorf("request = %s %s, want POST /emails", gotMethod, gotPath)
	}
	if gotAuth != "Bearer re_test_key" {
		t.Errorf("Authorization = %q, want Bearer re_test_key", gotAuth)
	}
	if gotCT != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", gotCT)
	}
	if gotBody.From != "info@karotammela.fi" ||
		len(gotBody.To) != 1 || gotBody.To[0] != "info@karotammela.fi" ||
		gotBody.ReplyTo != "visitor@example.com" ||
		gotBody.Subject != "karotammela.fi contact: Test Person" ||
		gotBody.Text != "Name: Test Person\nEmail: visitor@example.com\n\nHello there!" {
		t.Errorf("unexpected body: %+v", gotBody)
	}
}

// TestResendSendProviderError ensures non-2xx responses surface as errors that
// carry only the status code — never the provider's response body, which may
// echo submitted user content and must stay out of logs.
func TestResendSendProviderError(t *testing.T) {
	for _, status := range []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusTooManyRequests, http.StatusInternalServerError} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
			w.Write([]byte(`{"message":"secret provider detail with visitor@example.com"}`))
		}))
		s := &ResendSender{APIKey: "re_test_key", BaseURL: srv.URL}
		err := s.Send(context.Background(), ContactMessage{From: "a@b.fi", To: "c@d.fi", Subject: "s", Text: "t"})
		if err == nil {
			t.Errorf("status %d: expected error, got nil", status)
		} else {
			if !strings.Contains(err.Error(), "status") {
				t.Errorf("status %d: error %q should mention the status", status, err)
			}
			if strings.Contains(err.Error(), "secret provider detail") || strings.Contains(err.Error(), "visitor@example.com") {
				t.Errorf("status %d: error leaks provider body: %q", status, err)
			}
		}
		srv.Close()
	}
}

// TestResendSendNetworkError covers unreachable hosts.
func TestResendSendNetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close() // immediately unreachable

	s := &ResendSender{APIKey: "re_test_key", BaseURL: srv.URL}
	if err := s.Send(context.Background(), ContactMessage{From: "a@b.fi", To: "c@d.fi", Subject: "s", Text: "t"}); err == nil {
		t.Error("expected network error, got nil")
	}
}
