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
	h.Logout(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
