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
	Host                string
	Port                string
	ConfigFile          string
	ChallengeTTL        time.Duration
	CookieDomain        string
	CookieSecure        bool
	AppSecret           string
	SSRURL              string
	SessionTTL          time.Duration
	ProfileFetchTimeout time.Duration
	ProfileCacheTTL     time.Duration
}

func (c *Config) Address() string {
	if c.Host != "" || c.Port != "" {
		return net.JoinHostPort(valueOrDefault(c.Host, ""), valueOrDefault(c.Port, "3000"))
	}

	u, err := url.Parse(c.AppURL)

	if err != nil {
		log.Fatal(err)
	}

	return u.Host
}

func Load() Config {
	appURL := valueOr("APP_URL", "http://127.0.0.1:3000")

	return Config{
		AppName:             valueOr("APP_NAME", "nostr-auth"),
		AppURL:              appURL,
		Host:                valueOr("HOST", defaultHost(appURL)),
		Port:                portValueOr("PORT", defaultPort(appURL, "3000")),
		ConfigFile:          valueOr("CONFIG_FILE", "config.json"),
		ChallengeTTL:        durationValueOr("CHALLENGE_TTL", 5*time.Minute),
		CookieDomain:        valueOr("COOKIE_DOMAIN", defaultCookieDomain(appURL)),
		CookieSecure:        boolValueOr("COOKIE_SECURE", false),
		AppSecret:           valueOr("APP_KEY", "dev-app-key-change-me"),
		SSRURL:              os.Getenv("SSR_URL"),
		SessionTTL:          durationValueOr("SESSION_TTL", 168*time.Hour),
		ProfileFetchTimeout: durationValueOr("NOSTR_PROFILE_FETCH_TIMEOUT", 5*time.Second),
		ProfileCacheTTL:     durationValueOr("NOSTR_PROFILE_CACHE_TTL", 24*time.Hour),
	}
}

func valueOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}

func defaultHost(appURL string) string {
	parsed, err := url.Parse(appURL)
	if err != nil {
		return ""
	}

	return parsed.Hostname()
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

func defaultPort(appURL, fallback string) string {
	parsed, err := url.Parse(appURL)
	if err != nil {
		return fallback
	}

	if port := parsed.Port(); port != "" {
		return port
	}

	if parsed.Scheme == "https" {
		return "443"
	}

	if parsed.Scheme == "http" {
		return "80"
	}

	return fallback
}

func portValueOr(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	if _, err := strconv.Atoi(value); err != nil {
		return fallback
	}

	return value
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
