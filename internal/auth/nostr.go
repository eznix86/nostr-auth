package auth

import (
	"errors"
	"fmt"

	nostrlib "fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
)

func ParseAllowedPubKey(value string) (*nostrlib.PubKey, error) {
	if value == "" {
		return nil, nil
	}

	if len(value) == 64 {
		pubkey, err := nostrlib.PubKeyFromHex(value)
		if err != nil {
			return nil, err
		}

		return &pubkey, nil
	}

	prefix, decoded, err := nip19.Decode(value)
	if err != nil {
		return nil, err
	}

	if prefix != "npub" {
		return nil, fmt.Errorf("expected npub, got %s", prefix)
	}

	pubkey, ok := decoded.(nostrlib.PubKey)
	if !ok {
		return nil, errors.New("invalid npub payload")
	}

	return &pubkey, nil
}

func VerifyAuthEvent(evt nostrlib.Event, expectedChallenge, expectedRelay string, allowed *nostrlib.PubKey) error {
	if err := verifyEventIntegrity(evt); err != nil {
		return err
	}

	if err := verifyChallengeTag(evt, expectedChallenge); err != nil {
		return err
	}

	if err := verifyRelayTag(evt, expectedRelay); err != nil {
		return err
	}

	return verifyAllowedPubKey(evt, allowed)
}

func verifyEventIntegrity(evt nostrlib.Event) error {
	if !evt.CheckID() {
		return errors.New("invalid event id")
	}

	if !evt.VerifySignature() {
		return errors.New("invalid event signature")
	}

	if evt.Kind != 22242 {
		return fmt.Errorf("unexpected event kind %d", evt.Kind)
	}

	return nil
}

func verifyChallengeTag(evt nostrlib.Event, expectedChallenge string) error {
	challengeTag := evt.Tags.Find("challenge")
	if len(challengeTag) < 2 || challengeTag[1] != expectedChallenge {
		return errors.New("challenge mismatch")
	}

	return nil
}

func verifyRelayTag(evt nostrlib.Event, expectedRelay string) error {
	relayTag := evt.Tags.Find("relay")
	if len(relayTag) < 2 || relayTag[1] != expectedRelay {
		return errors.New("relay tag mismatch")
	}

	return nil
}

func verifyAllowedPubKey(evt nostrlib.Event, allowed *nostrlib.PubKey) error {
	if allowed != nil && evt.PubKey != *allowed {
		return errors.New("pubkey not allowed")
	}

	return nil
}
