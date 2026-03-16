package controller

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

	profile := h.LoadProfile(w, r, pubkey)
	host := ForwardedHost(r)
	if h.Authz == nil || !h.Authz.Allowed(host, pubkey) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	groups := h.Authz.Groups(host, pubkey)
	for key, value := range h.Nostr.Headers(pubkey, profile, groups) {
		w.Header().Set(key, value)
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "ok")
}
