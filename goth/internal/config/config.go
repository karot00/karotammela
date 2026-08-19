package config

import (
	"encoding/hex"
	"errors"
	"os"
	"strconv"
	"strings"

	"goth/internal/security"
)

type Config struct {
	Port               string
	Host               string
	DBPath             string
	GoogleAPIKey       string
	AIModel            string
	UnlockCookieSecret string
	BaseURL            string
	NextURL            string
	Env                string
	NextPingURL        string
	ResendAPIKey       string
	ContactFromEmail   string
	ContactToEmail     string
	CronSecret         string
	HNBaseURL          string
	GitHubBaseURL      string
	YahooBaseURL       string
	GeminiBaseURL      string
	BackupDir          string
	BackupKeep         int

	// VIP portal (MeetingPackage application). Disabled by default; one flag
	// closes every VIP route and hides every VIP link in both stacks.
	VIPEnabled      bool
	VIPPasswordHash string
	VIPCookieSecret string
	VIPAIModel      string
	VIPCVPath       string
	VIPContactEmail string
	VIPContactPhone string
	VIPContentDir   string
}

func Load() *Config {
	c := &Config{
		Port:               getEnv("GOTH_PORT", "8080"),
		Host:               getEnv("GOTH_HOST", "127.0.0.1"),
		DBPath:             getEnv("GOTH_DB_PATH", "goth.db"),
		GoogleAPIKey:       os.Getenv("GOOGLE_GENERATIVE_AI_API_KEY"),
		AIModel:            getEnv("AI_MODEL", "gemini-3.1-flash-lite"),
		UnlockCookieSecret: os.Getenv("UNLOCK_COOKIE_SECRET"),
		BaseURL:            getEnv("APP_URL", "https://karotammela.fi"),
		NextURL:            getEnv("NEXT_URL", "https://next.karotammela.fi"),
		Env:                getEnv("GOTH_ENV", "development"),
		NextPingURL:        os.Getenv("NEXT_PING_URL"),
		ResendAPIKey:       os.Getenv("RESEND_API_KEY"),
		ContactFromEmail:   os.Getenv("CONTACT_FROM_EMAIL"),
		ContactToEmail:     os.Getenv("CONTACT_TO_EMAIL"),
		CronSecret:         os.Getenv("CRON_SECRET"),
		HNBaseURL:          os.Getenv("GOTH_HN_BASE_URL"),
		GitHubBaseURL:      os.Getenv("GOTH_GITHUB_BASE_URL"),
		YahooBaseURL:       os.Getenv("GOTH_YAHOO_BASE_URL"),
		GeminiBaseURL:      os.Getenv("GOTH_GEMINI_BASE_URL"),
		BackupDir:          getEnv("GOTH_BACKUP_DIR", "backups"),
		BackupKeep:         getEnvInt("GOTH_BACKUP_KEEP", 14),

		VIPEnabled:      getEnvBool("VIP_ENABLED", false),
		VIPPasswordHash: os.Getenv("VIP_PASSWORD_HASH"),
		VIPCookieSecret: os.Getenv("VIP_COOKIE_SECRET"),
		VIPAIModel:      getEnv("VIP_AI_MODEL", "gemini-3.5-flash-lite"),
		VIPCVPath:       os.Getenv("VIP_CV_PATH"),
		VIPContactEmail: os.Getenv("VIP_CONTACT_EMAIL"),
		VIPContactPhone: os.Getenv("VIP_CONTACT_PHONE"),
		VIPContentDir:   getEnv("VIP_CONTENT_DIR", "content/vip"),
	}
	return c
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return fallback
	}
	return n
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	v, err := strconv.ParseBool(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return v
}

func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

// ContactConfigured reports whether contact email delivery is fully
// configured (all three values required, mirroring getContactConfig in the
// Next.js reference). Anything less disables delivery (503).
func (c *Config) ContactConfigured() bool {
	return c.ResendAPIKey != "" && c.ContactFromEmail != "" && c.ContactToEmail != ""
}

// VIPConfigured reports whether the VIP portal has the minimum credentials to
// run its access flow. VIPCVPath is optional: when unset, the CV download link
// and route stay hidden.
func (c *Config) VIPConfigured() bool {
	return c.VIPPasswordHash != "" && c.VIPCookieSecret != ""
}

// VIPStartupError fails startup when the VIP portal is enabled but too
// incomplete to run safely, so a half-configured portal fails closed instead
// of exposing a broken login. Enforced in production only; development may
// enable the routing skeleton without credentials.
func (c *Config) VIPStartupError() error {
	if c.Env != "development" && c.Env != "test" && c.Env != "production" {
		return errors.New("GOTH_ENV must be development, test, or production")
	}
	if c.IsProduction() && c.Host != "" && c.Host != "127.0.0.1" && c.Host != "::1" {
		return errors.New("production GOTH_HOST must be loopback")
	}
	if !c.VIPEnabled || !c.IsProduction() {
		return nil
	}
	if c.VIPPasswordHash == "" {
		return errors.New("VIP_ENABLED=true requires VIP_PASSWORD_HASH (Argon2id/scrypt hash, never plaintext)")
	}
	if !security.ValidateVIPPasswordHash(c.VIPPasswordHash, true) {
		return errors.New("VIP_PASSWORD_HASH is malformed or below the production work profile")
	}
	if c.VIPCookieSecret == "" {
		return errors.New("VIP_ENABLED=true requires VIP_COOKIE_SECRET (independent signing secret)")
	}
	if len(c.VIPCookieSecret) != 64 {
		return errors.New("VIP_COOKIE_SECRET must be exactly 64 hexadecimal characters")
	}
	if _, err := hex.DecodeString(c.VIPCookieSecret); err != nil {
		return errors.New("VIP_COOKIE_SECRET must be exactly 64 hexadecimal characters")
	}
	if strings.TrimSpace(c.VIPContentDir) == "" {
		return errors.New("VIP_ENABLED=true requires VIP_CONTENT_DIR")
	}
	return nil
}

func (c *Config) IntPort() int {
	p, err := strconv.Atoi(c.Port)
	if err != nil {
		return 8080
	}
	return p
}
