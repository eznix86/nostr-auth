package server

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	AuthErrorMissingChallenge     = "missing_challenge"
	AuthErrorInvalidPayload       = "invalid_payload"
	AuthErrorInvalidEvent         = "invalid_event"
	AuthErrorVerificationFailed   = "verification_failed"
	AuthErrorChallengeUnavailable = "challenge_unavailable"
	AuthErrorInvalidCSRF          = "invalid_csrf"
)

var DefaultRelays = []string{"wss://nos.lol", "wss://relay.damus.io", "wss://nostr.wine"}

type VerifyChallengeRequest struct {
	Event      string `json:"event"`
	RedirectTo string `json:"redirectTo"`
}

func SafeRedirect(target string) string {
	if target == "" {
		return "/"
	}

	parsed, err := url.Parse(target)
	if err != nil {
		return "/"
	}

	if parsed.Scheme == "" && parsed.Host == "" && strings.HasPrefix(parsed.Path, "/") {
		return target
	}

	if parsed.Scheme == "http" || parsed.Scheme == "https" {
		return target
	}

	return "/"
}

func ForwardedURL(r *http.Request) string {
	proto := headerOr(r, "X-Forwarded-Proto", "http")
	host := ForwardedHost(r)
	uri := r.Header.Get("X-Forwarded-Uri")

	if uri == "" {
		uri = "/"
	}

	return fmt.Sprintf("%s://%s%s", proto, host, uri)
}

func ForwardedHost(r *http.Request) string {
	host := r.Header.Get("X-Forwarded-Host")
	if host != "" {
		return host
	}

	return r.Host
}

func WantsBrowserRedirect(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "text/html")
}

func IsForwardAuthSubrequest(r *http.Request) bool {
	return r.Header.Get("X-Original-URI") != "" || r.Header.Get("X-Original-URL") != "" || r.Header.Get("X-Original-Method") != ""
}

func headerOr(r *http.Request, key, fallback string) string {
	value := r.Header.Get(key)
	if value == "" {
		return fallback
	}

	return value
}
