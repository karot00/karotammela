// genvectors appends created-by-Go golden vectors to the shared
// unlock-cookie vector file (Phase 12.5f). The Node side generates its
// vectors with scripts/generate-unlock-cookie-vectors.ts; this tool mints
// cookies with the real Go implementation (internal/security) and adds them
// so the Node test suite must reproduce and accept them, and vice versa.
//
// Run from the goth module:  go run ./test/genvectors
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"goth/internal/security"
)

const vectorsPath = "../shared/security/unlock-cookie-vectors.json"
const secret = "golden-vector-secret-0123456789abcdef"

type payload struct {
	SessionID  string `json:"sessionId"`
	Locale     string `json:"locale"`
	UnlockedAt int64  `json:"unlockedAt"`
}

type vector struct {
	Name         string   `json:"name"`
	CreatedBy    string   `json:"createdBy"`
	Secret       string   `json:"secret"`
	VerifySecret string   `json:"verifySecret,omitempty"`
	Payload      *payload `json:"payload"`
	Value        string   `json:"value"`
	ExpectVerify bool     `json:"expectVerify"`
	Note         string   `json:"note,omitempty"`
}

type doc struct {
	Description string          `json:"description"`
	Spec        json.RawMessage `json:"spec"`
	Vectors     []vector        `json:"vectors"`
}

func main() {
	data, err := os.ReadFile(vectorsPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read:", err)
		os.Exit(1)
	}
	var d doc
	if err := json.Unmarshal(data, &d); err != nil {
		fmt.Fprintln(os.Stderr, "parse:", err)
		os.Exit(1)
	}

	// Drop any previous Go vectors so re-runs are idempotent.
	kept := d.Vectors[:0]
	for _, v := range d.Vectors {
		if v.CreatedBy != "go" {
			kept = append(kept, v)
		}
	}
	d.Vectors = kept

	goVectors := []struct {
		name    string
		payload payload
		note    string
	}{
		{
			name:    "go-basic",
			payload: payload{SessionID: "sess-go-42", Locale: "en", UnlockedAt: time.Now().UnixMilli()},
			note:    "minted by internal/security.CreateUnlockCookieValue; Node must reproduce and accept",
		},
		{
			name:    "go-unicode",
			payload: payload{SessionID: "sess-åö-🚀-<&>", Locale: "fi", UnlockedAt: time.Now().UnixMilli()},
			note:    "multibyte + HTML-escape chars; Go marshals with SetEscapeHTML(false) for JSON.stringify byte parity",
		},
	}
	for _, gv := range goVectors {
		p := gv.payload
		d.Vectors = append(d.Vectors, vector{
			Name:      gv.name,
			CreatedBy: "go",
			Secret:    secret,
			Payload:   &p,
			Value: security.CreateUnlockCookieValue(security.UnlockPayload{
				SessionID:  p.SessionID,
				Locale:     p.Locale,
				UnlockedAt: p.UnlockedAt,
			}, secret),
			ExpectVerify: true,
			Note:         gv.note,
		})
	}

	out, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "marshal:", err)
		os.Exit(1)
	}
	if err := os.WriteFile(vectorsPath, append(out, '\n'), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		os.Exit(1)
	}
	fmt.Printf("appended %d Go vectors (%d total) to %s\n", len(goVectors), len(d.Vectors), vectorsPath)
}
