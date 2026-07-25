package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port               string
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
}

func Load() *Config {
	c := &Config{
		Port:               getEnv("GOTH_PORT", "8080"),
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

func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

// ContactConfigured reports whether contact email delivery is fully
// configured (all three values required, mirroring getContactConfig in the
// Next.js reference). Anything less disables delivery (503).
func (c *Config) ContactConfigured() bool {
	return c.ResendAPIKey != "" && c.ContactFromEmail != "" && c.ContactToEmail != ""
}

func (c *Config) IntPort() int {
	p, err := strconv.Atoi(c.Port)
	if err != nil {
		return 8080
	}
	return p
}
