package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"goth/internal/ai"
	"goth/internal/security"
)

func TestVIPChatRequiresCookieAndStreamsGroundedResponse(t *testing.T) {
	h := vipAccessHandlers(t, nil)
	h.cfg.GoogleAPIKey = "test-key"
	h.cfg.VIPAIModel = "vip-test-model"
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, "data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"Grounded answer\"}]}}]}\n\n")
	}))
	defer provider.Close()
	h.SetVIPGemini(&ai.GeminiStreamer{APIKey: "test-key", Model: "vip-test-model", Client: &http.Client{Transport: rewriteTransport{base: provider.URL, next: provider.Client().Transport}}})

	body := `{"messages":[{"role":"user","content":"What has Karo shipped?"}]}`
	unauth := httptest.NewRequest(http.MethodPost, "/api/vip/chat", strings.NewReader(body))
	unauth.Header.Set("Content-Type", "application/json")
	unauthRec := httptest.NewRecorder()
	h.VIPChat(unauthRec, unauth)
	if unauthRec.Code != http.StatusNotFound {
		t.Fatalf("unauthenticated status=%d", unauthRec.Code)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/vip/chat", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: security.VIPCookieName, Value: security.CreateVIPCookieValue(h.cfg.VIPCookieSecret, timeNow())})
	rec := httptest.NewRecorder()
	h.VIPChat(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Grounded answer") || !strings.Contains(rec.Body.String(), "event: done") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func timeNow() time.Time { return time.Now() }

type rewriteTransport struct {
	base string
	next http.RoundTripper
}

func (t rewriteTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	clone := r.Clone(r.Context())
	clone.URL.Scheme = "http"
	clone.URL.Host = strings.TrimPrefix(strings.TrimPrefix(t.base, "http://"), "https://")
	if t.next == nil {
		t.next = http.DefaultTransport
	}
	return t.next.RoundTrip(clone)
}
