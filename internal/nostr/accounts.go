package nostr

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

type Client struct {
	defaultRelays []string
	timeout       time.Duration
	cacheTTL      time.Duration
	fetchProfile  func(context.Context, string, []string, time.Duration) (*Profile, error)
	mu            sync.RWMutex
	cache         map[string]cachedProfile
}

type cachedProfile struct {
	profile   *Profile
	expiresAt time.Time
}

func NewClient(defaultRelays []string, timeout, cacheTTL time.Duration) *Client {
	return &Client{
		defaultRelays: append([]string(nil), defaultRelays...),
		timeout:       timeout,
		cacheTTL:      cacheTTL,
		fetchProfile:  FetchProfileFromRelays,
		cache:         make(map[string]cachedProfile),
	}
}

func (a *Client) FetchProfile(ctx context.Context, pubkey string) (*Profile, error) {
	if profile := a.profileFromCache(pubkey); profile != nil {
		return profile, nil
	}

	profile, err := a.fetchProfile(ctx, pubkey, a.defaultRelays, a.timeout)
	if err != nil {
		return nil, err
	}

	a.storeProfile(pubkey, profile)
	return profile, nil
}

func (a *Client) CachedProfile(pubkey string) *Profile {
	return a.profileFromCache(pubkey)
}

func (a *Client) Headers(pubkey string, profile *Profile, groups []string) map[string]string {
	npub := Npub(pubkey)
	headers := map[string]string{
		"Remote-User":         npub,
		"X-Forwarded-User":    npub,
		"X-Auth-Request-User": npub,
	}

	if profile == nil {
		if len(groups) > 0 {
			joinedGroups := strings.Join(groups, ",")
			headers["X-Forwarded-Groups"] = joinedGroups
			headers["X-Auth-Request-Groups"] = joinedGroups
		}
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
	if len(groups) > 0 {
		joinedGroups := strings.Join(groups, ",")
		headers["X-Forwarded-Groups"] = joinedGroups
		headers["X-Auth-Request-Groups"] = joinedGroups
	}

	return headers
}

func (a *Client) LoginURL(appURL, intendedURL string) string {
	return fmt.Sprintf("%s/?redirect=%s", strings.TrimRight(appURL, "/"), intendedURL)
}

func (a *Client) profileFromCache(pubkey string) *Profile {
	if pubkey == "" || a.cacheTTL <= 0 {
		return nil
	}

	a.mu.RLock()
	entry, ok := a.cache[pubkey]
	a.mu.RUnlock()
	if !ok {
		return nil
	}

	if time.Now().After(entry.expiresAt) {
		a.mu.Lock()
		delete(a.cache, pubkey)
		a.mu.Unlock()
		return nil
	}

	return entry.profile
}

func (a *Client) storeProfile(pubkey string, profile *Profile) {
	if pubkey == "" || profile == nil || a.cacheTTL <= 0 {
		return
	}

	a.mu.Lock()
	a.cache[pubkey] = cachedProfile{
		profile:   profile,
		expiresAt: time.Now().Add(a.cacheTTL),
	}
	a.mu.Unlock()
}
