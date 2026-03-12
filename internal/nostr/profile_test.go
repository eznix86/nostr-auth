package nostr

import (
	"errors"
	"testing"

	nostrlib "fiatjaf.com/nostr"
)

func TestResolveProfileRelaysFallsBackToDefaultsOnError(t *testing.T) {
	defaultRelays := []string{"wss://nos.lol", "wss://relay.damus.io"}
	relays := resolveProfileRelays(nil, errors.New("boom"), defaultRelays)

	if len(relays) != len(defaultRelays) {
		t.Fatalf("resolveProfileRelays() length = %d, want %d", len(relays), len(defaultRelays))
	}

	for i := range defaultRelays {
		if relays[i] != defaultRelays[i] {
			t.Fatalf("resolveProfileRelays()[%d] = %q, want %q", i, relays[i], defaultRelays[i])
		}
	}
}

func TestResolveProfileRelaysFallsBackWhenRelayListEmpty(t *testing.T) {
	defaultRelays := []string{"wss://nos.lol"}
	relays := resolveProfileRelays(nil, nil, defaultRelays)

	if len(relays) != 1 || relays[0] != defaultRelays[0] {
		t.Fatalf("resolveProfileRelays() = %v, want %v", relays, defaultRelays)
	}
}

func TestHasProfileDetailsAcceptsDisplayNameOnlyProfile(t *testing.T) {
	profile := &Profile{DisplayName: "bruno"}

	if !hasProfileDetails(profile) {
		t.Fatal("hasProfileDetails() = false, want true")
	}
}

func TestProfileFromEventRejectsNilEvent(t *testing.T) {
	secretKey := nostrlib.Generate()
	pubkey := secretKey.Public().Hex()

	profile, err := profileFromEvent(pubkey, secretKey.Public(), nil)
	if !errors.Is(err, ErrProfileEventNotFound) {
		t.Fatalf("profileFromEvent() error = %v, want %v", err, ErrProfileEventNotFound)
	}
	if profile != nil {
		t.Fatalf("profileFromEvent() = %v, want nil", profile)
	}
}
