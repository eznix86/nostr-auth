package server

import (
	"context"
	"encoding/json"
	"net/http"

	nostrlib "fiatjaf.com/nostr"
	"github.com/eznix86/nostr-auth/internal/challenge"
)

type Auth struct{ H *Context }

func (a *Auth) Verify(w http.ResponseWriter, r *http.Request) {
	ctx, ok := ChallengeFromContext(r)
	if !ok {
		a.H.AuthError(w, r, AuthErrorMissingChallenge)
		return
	}

	var payload VerifyChallengeRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		a.H.AuthError(w, r, AuthErrorInvalidPayload)
		return
	}

	var evt nostrlib.Event
	if err := json.Unmarshal([]byte(payload.Event), &evt); err != nil {
		a.H.AuthError(w, r, AuthErrorInvalidEvent)
		return
	}

	if err := a.H.NostrVerify.Event(evt, ctx.Challenge, r.Host); err != nil {
		a.H.Log.Error().Err(err).Str("handler", "auth.verify").Msg("failed to verify auth event")
		a.H.AuthError(w, r, AuthErrorVerificationFailed)
		return
	}

	if err := a.H.Challenge.Consume(ctx.SessionID, ctx.Challenge); err != nil {
		a.H.Log.Error().Err(err).Str("handler", "auth.verify").Msg("failed to consume challenge")
		a.H.AuthError(w, r, AuthErrorChallengeUnavailable)
		return
	}

	if !a.H.SetAuth(w, evt.PubKey.Hex()) {
		a.H.Fail(w, nil, "failed to issue auth session")
		return
	}

	profile, err := a.H.FetchProfile(context.Background(), evt.PubKey.Hex())
	if err != nil {
		a.H.Log.Error().Err(err).Str("handler", "auth.verify").Msg("failed to fetch profile")
	} else if profile != nil {
		a.H.Account.Set(w, profile)
	}

	a.H.Cookie.Clear(w, challenge.SessionCookieName)
	a.H.Redirect(w, r, SafeRedirect(payload.RedirectTo), http.StatusSeeOther)
}

func (a *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	a.H.ClearAuth(w)
	a.H.Cookie.Clear(w, challenge.SessionCookieName)
	a.H.CSRF.Clear(w)
	a.H.Account.Clear(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
