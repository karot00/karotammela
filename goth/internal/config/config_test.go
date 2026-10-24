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

func TestStartupError(t *testing.T) {
	cases := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{"production on loopback", Config{Env: "production", Host: "127.0.0.1"}, false},
		{"production on IPv6 loopback", Config{Env: "production", Host: "::1"}, false},
		{"production on a public address", Config{Env: "production", Host: "0.0.0.0"}, true},
		{"development on any address", Config{Env: "development", Host: "0.0.0.0"}, false},
		{"unknown environment", Config{Env: "staging"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.cfg.StartupError(); (err != nil) != tc.wantErr {
				t.Errorf("StartupError() = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
