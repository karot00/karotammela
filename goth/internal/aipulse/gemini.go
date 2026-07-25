package aipulse

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

// TextSummarizer generates a short completion for one prompt. Implemented by
// GeminiSummarizer. A nil TextSummarizer passed to a fetcher selects the
// deterministic fallback content (the source title/description), which is the
// offline behavior when GOOGLE_GENERATIVE_AI_API_KEY is unset.
type TextSummarizer interface {
	Generate(ctx context.Context, prompt string, maxOutputTokens int) (string, error)
}

// GeminiSummarizer is the non-streaming Gemini counterpart of ai.GeminiStreamer,
// mirroring the Vercel AI SDK `generateText` calls the Next.js fetchers use
// (plain prompt, no system instruction, bounded maxOutputTokens).
type GeminiSummarizer struct {
	APIKey string
	Model  string
	Client *http.Client
	// BaseURL overrides the API endpoint for tests; empty uses the real host.
	BaseURL string
}

const geminiGenerateEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent"

type generateContentRequest struct {
	Contents         []geminiContent `json:"contents"`
	GenerationConfig geminiGenConfig `json:"generationConfig"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenConfig struct {
	MaxOutputTokens int `json:"maxOutputTokens"`
}

type generateContentResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
}

// Generate returns the trimmed completion text. Errors are redacted: the
// provider body is discarded so prompts (story titles/URLs) never reach logs.
func (g *GeminiSummarizer) Generate(ctx context.Context, prompt string, maxOutputTokens int) (string, error) {
	client := g.Client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	endpoint := g.BaseURL
	if endpoint == "" {
		endpoint = fmt.Sprintf(geminiGenerateEndpoint, g.Model)
	}

	reqBody := generateContentRequest{
		Contents:         []geminiContent{{Role: "user", Parts: []geminiPart{{Text: prompt}}}},
		GenerationConfig: geminiGenConfig{MaxOutputTokens: maxOutputTokens},
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", g.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		// Discard the body: it can echo prompt content.
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("gemini status %d", resp.StatusCode)
	}
	var out generateContentResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&out); err != nil {
		return "", err
	}
	var sb strings.Builder
	for _, cand := range out.Candidates {
		for _, part := range cand.Content.Parts {
			sb.WriteString(part.Text)
		}
	}
	text := strings.TrimSpace(sb.String())
	if text == "" {
		return "", fmt.Errorf("gemini returned no text")
	}
	return text, nil
}
