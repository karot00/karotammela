package config

import "testing"

// TestIsProductionOnlyForExplicitProduction verifies the fail-closed env gate:
// missing or invalid GOTH_ENV values never enable production behavior (which
// would otherwise risk draft leakage or weaker cookie hardening).
func TestIsProductionOnlyForExplicitProduction(t *testing.T) {
	for _, tc := range []struct {
		env  string
		set  bool
		want bool
	}{
		{"", false, false},
		{"development", true, false},
		{"staging", true, false},
		{"PRODUCTION", true, false},
		{"production ", true, false},
		{"production", true, true},
	} {
		t.Setenv("GOTH_ENV", "")
		if tc.set {
			t.Setenv("GOTH_ENV", tc.env)
		} else {
			t.Setenv("GOTH_ENV", "")
		}
		c := Load()
		if c.IsProduction() != tc.want {
			t.Errorf("GOTH_ENV=%q -> IsProduction()=%v, want %v", tc.env, c.IsProduction(), tc.want)
		}
	}
}

// TestLoadDefaults confirms a sane non-production default and that required
// secrets remain empty until explicitly provided (never assumed present).
func TestLoadDefaults(t *testing.T) {
	t.Setenv("GOTH_ENV", "")
	t.Setenv("GOTH_PORT", "")
	t.Setenv("GOOGLE_GENERATIVE_AI_API_KEY", "")
	t.Setenv("UNLOCK_COOKIE_SECRET", "")
	c := Load()
	if c.IsProduction() {
		t.Error("default env must not be production")
	}
	if c.Port != "8080" {
		t.Errorf("default port = %q, want 8080", c.Port)
	}
	if c.GoogleAPIKey != "" || c.UnlockCookieSecret != "" {
		t.Error("secrets must default empty")
	}
}

func TestVIPDefaultsDisabled(t *testing.T) {
	t.Setenv("VIP_ENABLED", "")
	t.Setenv("VIP_PASSWORD_HASH", "")
	t.Setenv("VIP_COOKIE_SECRET", "")
	t.Setenv("VIP_AI_MODEL", "")
	t.Setenv("VIP_CV_PATH", "")
	t.Setenv("VIP_CONTACT_EMAIL", "")
	t.Setenv("VIP_CONTACT_PHONE", "")

	c := Load()
	if c.VIPEnabled {
		t.Error("VIPEnabled default = true, want false")
	}
	if c.VIPAIModel != "gemini-3.5-flash-lite" {
		t.Errorf("VIPAIModel default = %q, want gemini-3.5-flash-lite", c.VIPAIModel)
	}
	if c.VIPConfigured() {
		t.Error("VIPConfigured() = true with empty secrets, want false")
	}
	if err := c.VIPStartupError(); err != nil {
		t.Errorf("VIPStartupError() with VIP disabled = %v, want nil", err)
	}
}

func TestVIPEnvParsing(t *testing.T) {
	t.Setenv("VIP_ENABLED", "true")
	t.Setenv("VIP_PASSWORD_HASH", "$argon2id$v=19$m=65536,t=3,p=2$abc")
	t.Setenv("VIP_COOKIE_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("VIP_AI_MODEL", "gemini-3.1-flash-lite")
	t.Setenv("VIP_CV_PATH", "/var/lib/goth/private/cv.pdf")
	t.Setenv("VIP_CONTACT_EMAIL", "karo@example.com")
	t.Setenv("VIP_CONTACT_PHONE", "+358 400 234 711")

	c := Load()
	if !c.VIPEnabled {
		t.Error("VIPEnabled = false, want true")
	}
	if !c.VIPConfigured() {
		t.Error("VIPConfigured() = false with secrets set, want true")
	}
	if c.VIPAIModel != "gemini-3.1-flash-lite" {
		t.Errorf("VIPAIModel = %q, want override honored", c.VIPAIModel)
	}
	if c.VIPCVPath != "/var/lib/goth/private/cv.pdf" {
		t.Errorf("VIPCVPath = %q, want override honored", c.VIPCVPath)
	}
	if c.VIPContactEmail != "karo@example.com" {
		t.Errorf("VIPContactEmail = %q, want override honored", c.VIPContactEmail)
	}
	if c.VIPContactPhone != "+358 400 234 711" {
		t.Errorf("VIPContactPhone = %q, want override honored", c.VIPContactPhone)
	}
}

func TestVIPEnabledInvalidValueFallsBackToDisabled(t *testing.T) {
	t.Setenv("VIP_ENABLED", "not-a-bool")
	c := Load()
	if c.VIPEnabled {
		t.Error("VIPEnabled with unparsable value = true, want false (fail closed)")
	}
}

func TestVIPStartupErrorProduction(t *testing.T) {
	cases := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name:    "disabled production is always fine",
			cfg:     Config{Env: "production", VIPEnabled: false},
			wantErr: false,
		},
		{
			name:    "enabled production without hash fails",
			cfg:     Config{Env: "production", VIPEnabled: true, VIPCookieSecret: "secret"},
			wantErr: true,
		},
		{
			name:    "enabled production without cookie secret fails",
			cfg:     Config{Env: "production", VIPEnabled: true, VIPPasswordHash: "hash"},
			wantErr: true,
		},
		{
			name: "enabled production fully configured passes",
			cfg: Config{
				Env:             "production",
				VIPEnabled:      true,
				VIPPasswordHash: "$argon2id$v=19$m=65536,t=3,p=1$YWJjZGVmZ2hpamtsbW5vcA$YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXl6MDEyMzQ1",
				VIPCookieSecret: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				VIPContentDir:   "/var/lib/goth/private/vip",
			},
			wantErr: false,
		},
		{
			name:    "enabled development without secrets is allowed (skeleton work)",
			cfg:     Config{Env: "development", VIPEnabled: true},
			wantErr: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.VIPStartupError()
			if (err != nil) != tc.wantErr {
				t.Errorf("VIPStartupError() = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
