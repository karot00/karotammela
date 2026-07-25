package ai

import "testing"

func TestClampLevel(t *testing.T) {
	cases := []struct {
		in, want int
	}{
		{-5, 0}, {0, 0}, {50, 50}, {100, 100}, {120, 100},
	}
	for _, c := range cases {
		if got := clampLevel(c.in); got != c.want {
			t.Errorf("clampLevel(%d) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestExtractLevelTag(t *testing.T) {
	if got := ExtractLevelTag("hello [LEVEL:42]"); got == nil || *got != 42 {
		t.Errorf("ExtractLevelTag = %v, want 42", got)
	}
	if got := ExtractLevelTag("no tag here"); got != nil {
		t.Errorf("ExtractLevelTag = %v, want nil", got)
	}
	if got := ExtractLevelTag("over [LEVEL:140]"); got == nil || *got != 100 {
		t.Errorf("ExtractLevelTag clamp = %v, want 100", got)
	}
	if got := ExtractLevelTag("[LEVEL:abc]"); got != nil {
		t.Errorf("ExtractLevelTag bad = %v, want nil", got)
	}
}

func TestStripLevelTag(t *testing.T) {
	got := StripLevelTag("Welcome in. [LEVEL:99]")
	if got != "Welcome in." {
		t.Errorf("StripLevelTag = %q, want %q", got, "Welcome in.")
	}
}

func TestIsUnlockTriggered(t *testing.T) {
	if !IsUnlockTriggered("", 100) {
		t.Error("level 100 should trigger unlock")
	}
	if IsUnlockTriggered("", 99) {
		t.Error("level 99 should not trigger unlock")
	}
	if !IsUnlockTriggered("here is PROTOCOL_K_2026 yes", 0) {
		t.Error("access code should trigger unlock")
	}
}

func TestResolveNextLevel(t *testing.T) {
	// no parsed tag -> current + 5
	if got := ResolveNextLevel(10, nil); got != 15 {
		t.Errorf("ResolveNextLevel(nil) = %d, want 15", got)
	}
	// within cap
	parsed := 30
	if got := ResolveNextLevel(10, &parsed); got != 30 {
		t.Errorf("ResolveNextLevel = %d, want 30", got)
	}
	// above cap -> capped delta
	big := 100
	if got := ResolveNextLevel(10, &big); got != 40 {
		t.Errorf("ResolveNextLevel cap = %d, want 40", got)
	}
	// below floor -> -cap
	neg := 0
	if got := ResolveNextLevel(50, &neg); got != 20 {
		t.Errorf("ResolveNextLevel floor = %d, want 20", got)
	}
}

func TestGetInputLevelAdjustment(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"", -12},
		{"you idiot", -18},
		{"hi", -8},
		{"api security architecture workflow deploy pipeline because", 8},
		{"this is a reasonably long sentence about the build", 3},
		{"a short reply", 0},
	}
	for _, c := range cases {
		if got := GetInputLevelAdjustment(c.in); got != c.want {
			t.Errorf("GetInputLevelAdjustment(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestApplyInputAdjustment(t *testing.T) {
	if got := ApplyInputAdjustment(50, ""); got != 38 {
		t.Errorf("ApplyInputAdjustment empty = %d, want 38", got)
	}
	if got := ApplyInputAdjustment(95, "api security architecture workflow deploy pipeline because scale"); got != 100 {
		t.Errorf("ApplyInputAdjustment clamp = %d, want 100", got)
	}
}

func TestAccessCode(t *testing.T) {
	if GetAccessCode() != AccessCode {
		t.Error("GetAccessCode mismatch")
	}
}
