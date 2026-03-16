package authorization

import (
	"net"
	"strings"
)

func normalizeDomain(domain string) string {
	return strings.ToLower(strings.TrimSpace(domain))
}

func normalizeNIP05(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func stripPort(host string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return ""
	}

	parsedHost, _, err := net.SplitHostPort(host)
	if err == nil {
		return parsedHost
	}

	return host
}
