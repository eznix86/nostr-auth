package controller

import (
	"net/http"

	gonertia "github.com/romsar/gonertia/v2"
)

func (h *Handler) HomeIndex(w http.ResponseWriter, r *http.Request) {
	pubkey := AuthenticatedPubkeyFromContext(r)
	intendedURL := SafeRedirect(r.URL.Query().Get("redirect"), h.Authz)
	if pubkey != "" {
		target := "/logout"
		if intendedURL != "/" {
			target = intendedURL
		}

		h.Redirect(w, r, target, http.StatusSeeOther)
		return
	}

	if intendedURL == "/" {
		intendedURL = ""
	}
	if !h.SetIntendedURL(w, intendedURL) {
		h.Fail(w, nil, "failed to store intended url")
		return
	}

	err := h.Render(w, r, "Home", gonertia.Props{
		"title":       h.Config.AppName,
		"intendedUrl": intendedURL,
	})
	if err != nil {
		h.Fail(w, err, "failed to render home page")
	}
}
