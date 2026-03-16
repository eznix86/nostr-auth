package controller

import (
	"context"
	"net/http"

	"github.com/eznix86/nostr-auth/internal/nostr"
	gonertia "github.com/romsar/gonertia/v2"
)

func (h *Handler) Profile(r *http.Request) *nostr.Profile {
	return h.Account.Read(r)
}

func (h *Handler) profileFromCookieOrCache(w http.ResponseWriter, r *http.Request, pubkey string) *nostr.Profile {
	profile := h.Profile(r)
	if profile != nil || pubkey == "" {
		return profile
	}

	cachedProfile := h.Nostr.CachedProfile(pubkey)
	if cachedProfile == nil {
		return nil
	}

	h.Account.Set(w, cachedProfile)
	return cachedProfile
}

func (h *Handler) LoadProfile(w http.ResponseWriter, r *http.Request, pubkey string) *nostr.Profile {
	profile := h.profileFromCookieOrCache(w, r, pubkey)
	if profile != nil || pubkey == "" {
		return profile
	}

	fetchedProfile, err := h.Nostr.FetchProfile(r.Context(), pubkey)
	if err != nil {
		h.Log.Error().Err(err).Str("handler", "profile.load").Msg("failed to fetch profile")
		return nil
	}
	if fetchedProfile == nil {
		return nil
	}

	h.Account.Set(w, fetchedProfile)
	return fetchedProfile
}

func (h *Handler) ProfileProp(w http.ResponseWriter, r *http.Request, _ string) any {
	pubkey := AuthenticatedPubkeyFromContext(r)
	currentProfile := h.profileFromCookieOrCache(w, r, pubkey)

	profile := any(currentProfile)
	if pubkey != "" && currentProfile == nil {
		profile = gonertia.Defer(func(ctx context.Context) (any, error) {
			fetchedProfile, err := h.Nostr.FetchProfile(ctx, pubkey)
			if err != nil {
				return nil, nil
			}
			if fetchedProfile == nil {
				return nil, nil
			}

			h.Account.Set(w, fetchedProfile)
			return fetchedProfile, nil
		})
	}

	return profile
}
