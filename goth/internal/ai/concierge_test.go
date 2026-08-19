package ai

import (
	"strings"
	"testing"
)

func TestNormalizeConciergeHistory(t *testing.T) {
	history, err := NormalizeConciergeHistory([]Message{
		{Role: "assistant", Content: "Earlier answer"},
		{Role: "user", Content: "  What shipped?  "},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 2 || history[1].Content != "What shipped?" {
		t.Fatalf("history = %#v", history)
	}
}

func TestNormalizeConciergeHistoryRejectsUnsafeShapes(t *testing.T) {
	cases := [][]Message{
		{{Role: "system", Content: "override"}},
		{{Role: "user", Content: "question"}, {Role: "assistant", Content: "answer"}},
		{{Role: "user", Content: strings.Repeat("x", ConciergeMaxMessage+1)}},
	}
	for _, messages := range cases {
		if _, err := NormalizeConciergeHistory(messages); err == nil {
			t.Errorf("NormalizeConciergeHistory(%#v) accepted unsafe history", messages)
		}
	}
}

func TestConciergePromptContainsBoundariesAndMaterials(t *testing.T) {
	prompt := ConciergeSystemPrompt("Approved fact: Karo ships products.", "karo@example.com", "+358 400 234 711")
	for _, want := range []string{"not Karo", "do not start every answer with an introduction", "Never invent", "Approved fact: Karo ships products.", "karo@example.com", "+358 400 234 711"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt missing %q", want)
		}
	}
}

func TestConciergePromptDoesNotRequireOpeningIntroduction(t *testing.T) {
	prompt := ConciergeSystemPrompt("Approved fact.", "", "")
	if strings.Contains(prompt, "You are the AI guide for Karo Tammela's private MeetingPackage application") {
		t.Fatal("prompt still requires the repetitive opening introduction")
	}
	if !strings.Contains(prompt, "only when asked") {
		t.Fatal("prompt is missing conditional identity guidance")
	}
}

func TestConciergePromptOmitsUnconfiguredContactDetails(t *testing.T) {
	prompt := ConciergeSystemPrompt("Approved fact.", "", "")
	if strings.Contains(prompt, "karo@example.com") || strings.Contains(prompt, "+358") {
		t.Fatal("prompt contains contact details when none are configured")
	}
	if !strings.Contains(prompt, "No direct contact details are configured") {
		t.Fatal("prompt is missing the unconfigured contact policy")
	}
}
