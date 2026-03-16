package controller

import (
	"encoding/json"
	"net/http"

	nostrlib "fiatjaf.com/nostr"
	"github.com/eznix86/nostr-auth/internal/nostr"
)

type VerifyChallengeRequest struct {
	Event string `json:"event"`
}

func (h *Handler) VerifiedAuthPubkey(r *http.Request, challenge, host string) (string, string) {
	var payload VerifyChallengeRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return "", AuthErrorInvalidLogin
	}

	var evt nostrlib.Event
	if err := json.Unmarshal([]byte(payload.Event), &evt); err != nil {
		return "", AuthErrorInvalidLogin
	}

	if err := nostr.VerifyAuthEvent(evt, challenge, host); err != nil {
		h.Log.Error().Err(err).Str("handler", "auth.verify").Msg("failed to verify auth event")
		return "", AuthErrorInvalidLogin
	}

	return evt.PubKey.Hex(), ""
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
		h.AuthError(w, r, AuthErrorChallengeExpired)
		return
	}

	pubkey, authError := h.VerifiedAuthPubkey(r, ctx.Challenge, r.Host)
	if authError != "" {
		h.AuthError(w, r, authError)
		return
	}

	if !h.CompleteAuth(w, pubkey) {
		h.Fail(w, nil, "failed to issue auth session")
		return
	}

	h.Redirect(w, r, SafeRedirect(ctx.IntendedURL, h.Authz), http.StatusSeeOther)
}
