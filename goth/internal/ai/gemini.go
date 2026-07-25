package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const geminiEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse"

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiRequest struct {
	SystemInstruction *geminiContent  `json:"systemInstruction,omitempty"`
	Contents          []geminiContent `json:"contents"`
	GenerationConfig  geminiGenConfig `json:"generationConfig"`
}

type geminiGenConfig struct {
	Temperature     float64 `json:"temperature"`
	MaxOutputTokens int     `json:"maxOutputTokens"`
}

type geminiSSEData struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
}

// GeminiStreamer forwards Gemini SSE to a writer as text/event-stream.
type GeminiStreamer struct {
	APIKey string
	Model  string
	Client *http.Client
}

// StreamOptions configures a streaming request.
type StreamOptions struct {
	Locale  string
	History []Message
}

// Message mirrors the chat message shape.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Stream opens a Gemini SSE stream and writes token chunks to w using
// `event: token` SSE frames. The terminal `event: done` frame is emitted by the
// caller (handler) once the level is resolved. It returns the full concatenated
// text so the caller can resolve the level.
func (g *GeminiStreamer) Stream(opts StreamOptions, ctx context.Context, w io.Writer, flusher http.Flusher) (string, error) {
	if g.Client == nil {
		g.Client = &http.Client{Timeout: 90 * time.Second}
	}

	contents := make([]geminiContent, 0, len(opts.History))
	for _, m := range opts.History {
		role := m.Role
		if role == "assistant" {
			role = "model"
		}
		if role != "user" && role != "model" {
			role = "user"
		}
		contents = append(contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: m.Content}},
		})
	}

	reqBody := geminiRequest{
		SystemInstruction: &geminiContent{
			Role:  "user",
			Parts: []geminiPart{{Text: BuildSentinelSystemPrompt(opts.Locale)}},
		},
		Contents: contents,
		GenerationConfig: geminiGenConfig{
			Temperature:     0.7,
			MaxOutputTokens: 300,
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf(geminiEndpoint, g.Model)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", g.APIKey)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := g.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gemini error %d: %s", resp.StatusCode, string(body))
	}

	var full strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}
		var sse geminiSSEData
		if err := json.Unmarshal([]byte(data), &sse); err != nil {
			continue
		}
		for _, cand := range sse.Candidates {
			for _, part := range cand.Content.Parts {
				if part.Text == "" {
					continue
				}
				full.WriteString(part.Text)
				writeTokenEvent(w, part.Text)
				flusher.Flush()
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return full.String(), err
	}
	return full.String(), nil
}

func writeTokenEvent(w io.Writer, token string) {
	payload, _ := json.Marshal(token)
	fmt.Fprintf(w, "event: token\ndata: %s\n\n", payload)
}
