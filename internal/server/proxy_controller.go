package server

import (
	"io"
	"net/http"
	"net/url"
)

type Proxy struct{ H *Context }

func (p *Proxy) Check(w http.ResponseWriter, r *http.Request) {
	pubkey := AuthenticatedPubkeyFromContext(r)
	if pubkey == "" {
		if WantsBrowserRedirect(r) && !IsForwardAuthSubrequest(r) {
			target := p.H.NostrAccounts.LoginURL(p.H.Config.AppURL, url.QueryEscape(ForwardedURL(r)))
			http.Redirect(w, r, target, http.StatusTemporaryRedirect)
			return
		}

		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	profile := p.H.Profile(r)
	nip05 := ""
	if profile != nil {
		nip05 = profile.NIP05
	}
	if !p.H.Allowed(ForwardedHost(r), pubkey, nip05) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	for key, value := range p.H.NostrAccounts.Headers(pubkey, profile) {
		w.Header().Set(key, value)
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "ok")
}
