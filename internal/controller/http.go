package controller

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	AuthErrorChallengeExpired   = "Your login challenge expired. Please start again."
	AuthErrorInvalidLogin       = "We could not verify your Nostr login. Please try again."
	AuthErrorInvalidCSRF        = "Your session expired. Please try again."
	AuthErrorInvalidRedirect    = "We could not send you back to the requested app."
)

type redirectAuthorizer interface {
	AllowsRedirect(target string) bool
}

func SafeRedirect(target string, authz redirectAuthorizer) string {
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

	if authz != nil && authz.AllowsRedirect(target) {
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
