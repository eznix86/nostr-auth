package nostr

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Accounts struct {
	defaultRelays []string
	timeout       time.Duration
}

func NewAccounts(defaultRelays []string, timeout time.Duration) *Accounts {
	return &Accounts{defaultRelays: append([]string(nil), defaultRelays...), timeout: timeout}
}

func (a *Accounts) FetchProfile(ctx context.Context, pubkey string) (*Profile, error) {
	return FetchProfileFromRelays(ctx, pubkey, a.defaultRelays, a.timeout)
}

func (a *Accounts) Headers(pubkey string, profile *Profile) map[string]string {
	npub := Npub(pubkey)
	headers := map[string]string{
		"Remote-User":         npub,
		"X-Forwarded-User":    npub,
		"X-Auth-Request-User": npub,
	}

	if profile == nil {
		return headers
	}

	if profile.NIP05 != "" {
		headers["X-Auth-Request-Email"] = profile.NIP05
	}
	if profile.DisplayName != "" {
		headers["X-Auth-Request-Preferred-Username"] = profile.DisplayName
	}
	if profile.Name != "" {
		headers["Remote-User-Name"] = profile.Name
		headers["X-Auth-Request-Name"] = profile.Name
	}
	if profile.Picture != "" {
		headers["Remote-User-Picture"] = profile.Picture
		headers["X-Auth-Request-Picture"] = profile.Picture
	}

	return headers
}

func (a *Accounts) LoginURL(appURL, redirectTo string) string {
	return fmt.Sprintf("%s/?redirect=%s", strings.TrimRight(appURL, "/"), redirectTo)
}
