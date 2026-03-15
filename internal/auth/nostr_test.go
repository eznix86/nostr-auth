package auth

import (
	"testing"

	nostrlib "fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
)

func TestParseAllowedPubKeyFromNpub(t *testing.T) {
	secretKey := nostrlib.Generate()
	want := secretKey.Public()

	pubkey, err := ParseAllowedPubKey(nip19.EncodeNpub(want))
	if err != nil {
		t.Fatalf("ParseAllowedPubKey() error = %v", err)
	}

	if pubkey == nil || *pubkey != want {
		t.Fatalf("ParseAllowedPubKey() = %v, want %v", pubkey, want)
	}
}

func TestVerifyAuthEvent(t *testing.T) {
	secretKey := nostrlib.Generate()
	allowed := secretKey.Public()

	evt := nostrlib.Event{
		CreatedAt: nostrlib.Now(),
		Kind:      22242,
		Tags:      nostrlib.Tags{{"challenge", "challenge-123"}, {"relay", "127.0.0.1:3000"}},
	}

	if err := evt.Sign(secretKey); err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	if err := VerifyAuthEvent(evt, "challenge-123", "127.0.0.1:3000", &allowed); err != nil {
		t.Fatalf("VerifyAuthEvent() error = %v", err)
	}
}

func TestVerifyAuthEventRejectsWrongChallenge(t *testing.T) {
	secretKey := nostrlib.Generate()

	evt := nostrlib.Event{
		CreatedAt: nostrlib.Now(),
		Kind:      22242,
		Tags:      nostrlib.Tags{{"challenge", "challenge-123"}, {"relay", "127.0.0.1:3000"}},
	}

	if err := evt.Sign(secretKey); err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	if err := VerifyAuthEvent(evt, "challenge-456", "127.0.0.1:3000", nil); err == nil {
		t.Fatal("VerifyAuthEvent() expected error for wrong challenge, got nil")
	}
}
