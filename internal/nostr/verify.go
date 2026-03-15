package nostr

import (
	"errors"
	"fmt"

	nostrlib "fiatjaf.com/nostr"
	"fiatjaf.com/nostr/nip19"
)

func VerifyAuthEvent(evt nostrlib.Event, expectedChallenge, expectedRelay string) error {
	if !evt.CheckID() {
		return errors.New("invalid event id")
	}

	if !evt.VerifySignature() {
		return errors.New("invalid event signature")
	}

	if evt.Kind != 22242 {
		return fmt.Errorf("unexpected event kind %d", evt.Kind)
	}

	challengeTag := evt.Tags.Find("challenge")
	if len(challengeTag) < 2 || challengeTag[1] != expectedChallenge {
		return errors.New("challenge mismatch")
	}

	relayTag := evt.Tags.Find("relay")
	if len(relayTag) < 2 || relayTag[1] != expectedRelay {
		return errors.New("relay tag mismatch")
	}

	return nil
}

func Npub(pubkey string) string {
	pk, err := nostrlib.PubKeyFromHex(pubkey)
	if err != nil {
		return pubkey
	}

	return nip19.EncodeNpub(pk)
}
