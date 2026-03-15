package server

import (
	"io"
	"net/http"
	"net/url"
)

func (h *Handler) ProxyCheck(w http.ResponseWriter, r *http.Request) {
	pubkey := AuthenticatedPubkeyFromContext(r)
	if pubkey == "" {
		if WantsBrowserRedirect(r) && !IsForwardAuthSubrequest(r) {
			target := h.Nostr.LoginURL(h.Config.AppURL, url.QueryEscape(ForwardedURL(r)))
			http.Redirect(w, r, target, http.StatusTemporaryRedirect)
			return
		}

		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	profile := h.Profile(r)
	if profile == nil {
		fetched, err := h.FetchProfile(r.Context(), pubkey)
		if err != nil {
			h.Log.Error().Err(err).Str("handler", "proxy.check").Msg("failed to fetch profile")
		} else if fetched != nil {
			profile = fetched
			h.Account.Set(w, profile)
		}
	}

	nip05 := ""
	if profile != nil {
		nip05 = profile.NIP05
	}
	host := ForwardedHost(r)
	if !h.Allowed(host, pubkey, nip05) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	groups := h.Groups(host, pubkey, nip05)

	for key, value := range h.Nostr.Headers(pubkey, profile, groups) {
		w.Header().Set(key, value)
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "ok")
}
