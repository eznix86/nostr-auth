package server

import (
	"net/http"

	gonertia "github.com/romsar/gonertia/v2"
)

type Home struct{ H *Context }

func (h *Home) Index(w http.ResponseWriter, r *http.Request) {
	redirectTo := r.URL.Query().Get("redirect")
	if AuthenticatedPubkeyFromContext(r) != "" && redirectTo != "" {
		h.H.Redirect(w, r, SafeRedirect(redirectTo), http.StatusSeeOther)
		return
	}
	if AuthenticatedPubkeyFromContext(r) != "" {
		h.H.Redirect(w, r, "/logout", http.StatusSeeOther)
		return
	}

	err := h.H.Render(w, r, "Home", gonertia.Props{
		"title":      h.H.Config.AppName,
		"message":    "A tiny Nostr-first auth app in Go.",
		"redirectTo": redirectTo,
		"authError":  r.URL.Query().Get("auth_error"),
	})
	if err != nil {
		h.H.Fail(w, err, "failed to render home page")
	}
}
