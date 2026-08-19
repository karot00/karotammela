package ai

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type flushBuffer struct{ strings.Builder }

func (f *flushBuffer) Flush() {}

func TestStreamWithPromptUsesPromptAndParsesTokens(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "Approved application materials") {
			t.Error("request did not contain the concierge system prompt")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, "data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"Hello\"}]}}]}\n\n")
		io.WriteString(w, "data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\" there\"}]}}]}\n\n")
	}))
	defer server.Close()

	streamer := &GeminiStreamer{APIKey: "test", Model: "test", Client: server.Client()}
	old := geminiEndpoint
	geminiEndpoint = server.URL + "/models/%s"
	defer func() { geminiEndpoint = old }()

	var out flushBuffer
	full, err := streamer.StreamWithPrompt(StreamOptions{
		SystemPrompt: "Approved application materials",
		History:      []Message{{Role: "user", Content: "Hello?"}},
	}, context.Background(), &out, &out)
	if err != nil {
		t.Fatal(err)
	}
	if full != "Hello there" || !strings.Contains(out.String(), "event: token") {
		t.Fatalf("full=%q stream=%q", full, out.String())
	}
}
