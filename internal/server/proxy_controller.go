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
	if profile == nil {
		fetched, err := p.H.FetchProfile(r.Context(), pubkey)
		if err != nil {
			p.H.Log.Error().Err(err).Str("handler", "proxy.check").Msg("failed to fetch profile")
		} else if fetched != nil {
			profile = fetched
			p.H.Account.Set(w, profile)
		}
	}

	nip05 := ""
	if profile != nil {
		nip05 = profile.NIP05
	}
	host := ForwardedHost(r)
	if !p.H.Allowed(host, pubkey, nip05) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	groups := p.H.Groups(host, pubkey, nip05)

	for key, value := range p.H.NostrAccounts.Headers(pubkey, profile, groups) {
		w.Header().Set(key, value)
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "ok")
}
