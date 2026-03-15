package server

import (
	"net/http"

	gonertia "github.com/romsar/gonertia/v2"
)

func (h *Handler) HomeIndex(w http.ResponseWriter, r *http.Request) {
	intendedURL := SafeRedirect(r.URL.Query().Get("redirect"))
	if AuthenticatedPubkeyFromContext(r) != "" && intendedURL != "/" {
		h.Redirect(w, r, intendedURL, http.StatusSeeOther)
		return
	}
	if AuthenticatedPubkeyFromContext(r) != "" {
		h.Redirect(w, r, "/logout", http.StatusSeeOther)
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
