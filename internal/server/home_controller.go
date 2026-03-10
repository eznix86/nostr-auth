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

	sessionID, err := h.H.EnsureChallengeSession(w, r)
	if err != nil {
		h.H.Fail(w, err, "failed to issue challenge session")
		return
	}

	challenge, err := h.H.CurrentOrIssueChallenge(sessionID)
	if err != nil {
		h.H.Fail(w, err, "failed to generate challenge")
		return
	}

	csrfToken, err := h.H.CSRF.Ensure(w, r)
	if err != nil {
		h.H.Fail(w, err, "failed to issue csrf token")
		return
	}

	err = h.H.Render(w, r, "Home", gonertia.Props{
		"title":               h.H.Config.AppName,
		"message":             "A tiny Nostr-first auth app in Go.",
		"challenge":           challenge.Token,
		"csrfToken":           csrfToken,
		"relay":               r.Host,
		"redirectTo":          redirectTo,
		"authenticatedPubkey": AuthenticatedPubkeyFromContext(r),
		"authError":           r.URL.Query().Get("auth_error"),
		"profile":             gonertia.Defer(h.H.Profile(r)),
	})
	if err != nil {
		h.H.Fail(w, err, "failed to render home page")
	}
}
