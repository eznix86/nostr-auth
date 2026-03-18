package controller

import (
	"net/http"

	gonertia "github.com/romsar/gonertia/v2"
)

func (h *Handler) HomeIndex(w http.ResponseWriter, r *http.Request) {
	pubkey := AuthenticatedPubkeyFromContext(r)
	queryRedirect := r.URL.Query().Get("redirect")
	intendedURL := SafeRedirect(queryRedirect, h.Authz)
	if intendedURL == "/" {
		intendedURL = ""
	}

	h.Log.Info().
		Str("handler", "home.index").
		Str("pubkey", pubkey).
		Str("query_redirect", queryRedirect).
		Str("safe_redirect", intendedURL).
		Msg("home redirect context")

	if pubkey != "" {
		storedIntendedURL := ""
		if intendedURL == "" {
			storedIntendedURL = h.IntendedURL(r)
			intendedURL = storedIntendedURL
		}

		target := "/logout"
		if intendedURL != "" {
			target = intendedURL
		}

		h.Log.Info().
			Str("handler", "home.index").
			Str("pubkey", pubkey).
			Str("stored_intended_url", storedIntendedURL).
			Str("target", target).
			Msg("redirecting authenticated user from home")

		h.Redirect(w, r, target, http.StatusSeeOther)
		return
	}

	if !h.SetIntendedURL(w, intendedURL) {
		h.Fail(w, nil, "failed to store intended url")
		return
	}

	h.Log.Info().
		Str("handler", "home.index").
		Str("stored_intended_url", intendedURL).
		Msg("stored intended url for guest")

	err := h.Render(w, r, "Home", gonertia.Props{
		"title":       h.Config.AppName,
		"intendedUrl": intendedURL,
	})
	if err != nil {
		h.Fail(w, err, "failed to render home page")
	}
}
