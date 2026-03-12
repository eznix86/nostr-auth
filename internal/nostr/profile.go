package nostr

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
)

type Profile struct {
	PubKey      string `json:"pubkey"`
	NPub        string `json:"npub"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Picture     string `json:"picture"`
	Bio         string `json:"about"`
	Website     string `json:"website"`
	NIP05       string `json:"nip05"`
	Banner      string `json:"banner"`
}

type profileMetadata struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Picture     string `json:"picture"`
	Bio         string `json:"about"`
	Website     string `json:"website"`
	NIP05       string `json:"nip05"`
	Banner      string `json:"banner"`
}

var ErrProfileEventNotFound = errors.New("profile event not found")

var pool struct {
	instance *nostr.Pool
	once     sync.Once
}

// GetPool returns a shared Nostr pool instance (lazy initialized)
func GetPool() *nostr.Pool {
	pool.once.Do(func() {
		pool.instance = nostr.NewPool(nostr.PoolOptions{})
	})
	return pool.instance
}

// ClosePool closes the shared pool (useful for testing)
func ClosePool() {
	if pool.instance != nil {
		pool.instance.Close("pool closed")
		pool.instance = nil
	}
	// Reset once so a new pool can be created
	pool.once = sync.Once{}
}

func FetchUserRelays(ctx context.Context, pubKey string, relayURLs []string) ([]string, error) {
	if pubKey == "" {
		return nil, nil
	}

	pk, err := parsePubKey(pubKey)
	if err != nil {
		return nil, err
	}

	filter := nostr.Filter{
		Kinds:   []nostr.Kind{10002},
		Authors: []nostr.PubKey{pk},
	}

	event, err := fetchFirstEvent(ctx, relayURLs, filter)
	if err != nil {
		return nil, err
	}

	return extractRelayURLs(event), nil
}

func FetchProfile(ctx context.Context, pubKey string, relayURL string) (*Profile, error) {
	if pubKey == "" {
		return nil, nil
	}

	pk, err := parsePubKey(pubKey)
	if err != nil {
		return nil, err
	}

	filter := nostr.Filter{
		Kinds:   []nostr.Kind{0},
		Authors: []nostr.PubKey{pk},
	}

	event, err := fetchFirstEvent(ctx, []string{relayURL}, filter)
	if err != nil {
		return nil, err
	}

	return profileFromEvent(pubKey, pk, event)
}

func FetchProfileFromRelays(ctx context.Context, pubKey string, defaultRelays []string, timeout time.Duration) (*Profile, error) {
	relayListCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	relays, err := FetchUserRelays(relayListCtx, pubKey, defaultRelays)
	relays = resolveProfileRelays(relays, err, defaultRelays)

	for _, relay := range relays {
		relayCtx, relayCancel := context.WithTimeout(ctx, timeout)
		profile, err := FetchProfile(relayCtx, pubKey, relay)
		relayCancel()
		if err != nil {
			continue
		}
		if hasProfileDetails(profile) {
			return profile, nil
		}
	}

	if len(defaultRelays) == 0 {
		return nil, ErrProfileEventNotFound
	}

	finalCtx, finalCancel := context.WithTimeout(ctx, timeout)
	defer finalCancel()
	return FetchProfile(finalCtx, pubKey, defaultRelays[0])
}

func FetchProfileWithTimeout(pubKey string, relayURL string, timeout time.Duration) (*Profile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return FetchProfile(ctx, pubKey, relayURL)
}

func parsePubKey(pubKey string) (nostr.PubKey, error) {
	return nostr.PubKeyFromHex(pubKey)
}

func fetchFirstEvent(ctx context.Context, relayURLs []string, filter nostr.Filter) (*nostr.Event, error) {
	events := GetPool().FetchMany(ctx, relayURLs, filter, nostr.SubscriptionOptions{})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case re, ok := <-events:
		if !ok {
			return nil, ErrProfileEventNotFound
		}
		return &re.Event, nil
	}
}

func extractRelayURLs(event *nostr.Event) []string {
	if event == nil {
		return nil
	}

	var relays []string
	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "r" {
			relays = append(relays, tag[1])
		}
	}

	return relays
}

func profileFromEvent(pubKey string, pk nostr.PubKey, event *nostr.Event) (*Profile, error) {
	if event == nil {
		return nil, ErrProfileEventNotFound
	}

	var meta profileMetadata
	if err := json.Unmarshal([]byte(event.Content), &meta); err != nil {
		return nil, err
	}

	return &Profile{
		PubKey:      pubKey,
		NPub:        nip19.EncodeNpub(pk),
		Name:        meta.Name,
		DisplayName: meta.DisplayName,
		Picture:     meta.Picture,
		Bio:         meta.Bio,
		Website:     meta.Website,
		NIP05:       meta.NIP05,
		Banner:      meta.Banner,
	}, nil
}

func resolveProfileRelays(relays []string, err error, defaultRelays []string) []string {
	if err != nil || len(relays) == 0 {
		return defaultRelays
	}

	return relays
}

func hasProfileDetails(profile *Profile) bool {
	if profile == nil {
		return false
	}

	return profile.Name != "" || profile.DisplayName != "" || profile.Picture != "" || profile.NIP05 != ""
}
