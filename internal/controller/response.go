package controller

import (
	"net/http"

	gonertia "github.com/romsar/gonertia/v2"
)

func (h *Handler) Render(w http.ResponseWriter, r *http.Request, component string, props gonertia.Props) error {
	return h.Inertia.Render(w, r, component, props)
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request, target string, status int) {
	h.Inertia.Redirect(w, r, target, status)
}

func (h *Handler) AuthError(w http.ResponseWriter, r *http.Request, code string) {
	ctx := gonertia.SetFlash(r.Context(), gonertia.Flash{"error": code})
	h.Inertia.Redirect(w, r.WithContext(ctx), "/", http.StatusSeeOther)
}

func (h *Handler) Fail(w http.ResponseWriter, err error, message string) {
	h.Log.Error().Err(err).Msg(message)
	http.Error(w, "internal server error", http.StatusInternalServerError)
}
