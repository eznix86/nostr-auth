package config

import (
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppName             string
	AppURL              string
	AuthConfigFile      string
	ChallengeTTL        time.Duration
	CookieDomain        string
	CookieSecure        bool
	AppSecret           string
	SSRURL              string
	SessionTTL          time.Duration
	ProfileFetchTimeout time.Duration
}

func (c *Config) Address() string {
	u, err := url.Parse(c.AppURL)

	if err != nil {
		log.Fatal(err)
	}

	return u.Host
}

func Load() Config {
	return Config{
		AppName:             valueOr("APP_NAME", "nostr-auth"),
		AppURL:              valueOr("APP_URL", "http://127.0.0.1:3000"),
		AuthConfigFile:      valueOr("AUTH_CONFIG_FILE", "auth.json"),
		ChallengeTTL:        durationValueOr("CHALLENGE_TTL", 5*time.Minute),
		CookieDomain:        valueOr("COOKIE_DOMAIN", defaultCookieDomain(valueOr("APP_URL", "http://127.0.0.1:3000"))),
		CookieSecure:        boolValueOr("COOKIE_SECURE", false),
		AppSecret:           valueOr("APP_KEY", "dev-app-key-change-me"),
		SSRURL:              os.Getenv("SSR_URL"),
		SessionTTL:          durationValueOr("SESSION_TTL", 168*time.Hour),
		ProfileFetchTimeout: durationValueOr("NOSTR_PROFILE_FETCH_TIMEOUT", 5*time.Second),
	}
}

func valueOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func defaultCookieDomain(appURL string) string {
	parsed, err := url.Parse(appURL)
	if err != nil {
		return ""
	}

	host := strings.ToLower(parsed.Hostname())
	if host == "" || host == "localhost" || net.ParseIP(host) != nil {
		return ""
	}

	return host
}

func durationValueOr(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func boolValueOr(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}
