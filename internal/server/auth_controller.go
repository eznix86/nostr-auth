package server

import (
	"encoding/json"
	"net/http"

	nostrlib "fiatjaf.com/nostr"
	"github.com/eznix86/nostr-auth/internal/nostr"
	"github.com/eznix86/nostr-auth/internal/session"
)

type VerifyChallengeRequest struct {
	Event string `json:"event"`
}

func (h *Handler) AuthCSRF(w http.ResponseWriter, r *http.Request) {
	token, err := h.ensureCSRF(w, r)
	if err != nil {
		h.Fail(w, err, "failed to issue csrf token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (h *Handler) AuthChallenge(w http.ResponseWriter, r *http.Request) {
	token, err := h.IssueChallenge(w, h.IntendedURL(r))
	if err != nil {
		h.Fail(w, err, "failed to generate challenge")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"token": token,
		"relay": r.Host,
	})
}

func (h *Handler) AuthVerify(w http.ResponseWriter, r *http.Request) {
	ctx, ok := ChallengeFromContext(r)
	if !ok {
		h.AuthError(w, r, AuthErrorMissingChallenge)
		return
	}

	var payload VerifyChallengeRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.AuthError(w, r, AuthErrorInvalidPayload)
		return
	}

	var evt nostrlib.Event
	if err := json.Unmarshal([]byte(payload.Event), &evt); err != nil {
		h.AuthError(w, r, AuthErrorInvalidEvent)
		return
	}

	if err := nostr.VerifyAuthEvent(evt, ctx.Challenge, r.Host); err != nil {
		h.Log.Error().Err(err).Str("handler", "auth.verify").Msg("failed to verify auth event")
		h.AuthError(w, r, AuthErrorVerificationFailed)
		return
	}

	if !h.SetAuth(w, evt.PubKey.Hex()) {
		h.Fail(w, nil, "failed to issue auth session")
		return
	}

	h.Cookie.Clear(w, session.ChallengeCookieName)
	h.ClearIntendedURL(w)
	h.Redirect(w, r, SafeRedirect(ctx.IntendedURL), http.StatusSeeOther)
}
