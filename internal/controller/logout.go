package controller

import (
	"net/http"

	gonertia "github.com/romsar/gonertia/v2"
)

func (h *Handler) LogoutIndex(w http.ResponseWriter, r *http.Request) {
	pubkey := AuthenticatedPubkeyFromContext(r)
	if pubkey == "" {
		h.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	err := h.Render(w, r, "Logout", gonertia.Props{
		"title":               h.Config.AppName,
		"authenticatedPubkey": pubkey,
		"profile":             h.ProfileProp(w, r, "logout.index"),
	})
	if err != nil {
		h.Fail(w, err, "failed to render logout page")
	}
}

func (h *Handler) LogoutSubmit(w http.ResponseWriter, r *http.Request) {
	intendedURL := h.IntendedURL(r)
	if intendedURL == "" {
		intendedURL = SafeRedirect(r.URL.Query().Get("redirect"), h.Authz)
		if intendedURL == "/" {
			intendedURL = ""
		}
	}

	h.Logout(w)
	if intendedURL != "" {
		if !h.SetIntendedURL(w, intendedURL) {
			h.Fail(w, nil, "failed to store intended url")
			return
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
