package server

import (
	"context"
	"net/http"

	"github.com/eznix86/nostr-auth/internal/account"
	"github.com/eznix86/nostr-auth/internal/authorization"
	"github.com/eznix86/nostr-auth/internal/challenge"
	"github.com/eznix86/nostr-auth/internal/config"
	"github.com/eznix86/nostr-auth/internal/cookie"
	"github.com/eznix86/nostr-auth/internal/csrf"
	appinertia "github.com/eznix86/nostr-auth/internal/inertia"
	"github.com/eznix86/nostr-auth/internal/nostr"
	"github.com/eznix86/nostr-auth/internal/session"
	gonertia "github.com/romsar/gonertia/v2"
	"github.com/rs/zerolog"
)

type contextKey string

const (
	authenticatedPubkeyKey contextKey = "authenticated_pubkey"
	challengeContextKey    contextKey = "challenge_context"
)

type ChallengeContext struct {
	SessionID string
	Challenge string
}

type Context struct {
	Config        config.Config
	Log           zerolog.Logger
	Account       *account.Cookie
	Authz         *authorization.Policy
	Challenge     *challenge.Store
	Cookie        *cookie.Jar
	CSRF          *csrf.Guard
	Inertia       *appinertia.App
	NostrAccounts *nostr.Accounts
	NostrVerify   *nostr.Verify
	Session       *session.Signer
}

func (h *Context) Text(w http.ResponseWriter, status int, body string) {
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

func (h *Context) Render(w http.ResponseWriter, r *http.Request, component string, props gonertia.Props) error {
	return h.Inertia.Render(w, r, component, props)
}

func (h *Context) Redirect(w http.ResponseWriter, r *http.Request, target string, status int) {
	h.Inertia.Redirect(w, r, target, status)
}

func (h *Context) AuthError(w http.ResponseWriter, r *http.Request, code string) {
	http.Redirect(w, r, "/?auth_error="+code, http.StatusSeeOther)
}

func (h *Context) Fail(w http.ResponseWriter, err error, message string) {
	h.Log.Error().Err(err).Msg(message)
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

func (h *Context) AuthenticatedPubkey(r *http.Request) string {
	claims, ok := h.Claims(r)
	if !ok {
		return ""
	}

	return claims.PubKey
}

func (h *Context) Claims(r *http.Request) (*session.Claims, bool) {
	token := h.Cookie.Value(r, session.CookieName)
	if token == "" {
		return nil, false
	}

	claims, err := h.Session.Verify(token)
	if err != nil {
		return nil, false
	}

	return claims, true
}

func (h *Context) SetAuth(w http.ResponseWriter, pubkey string) bool {
	token, err := h.Session.Sign(pubkey)
	if err != nil {
		return false
	}

	h.Cookie.Set(w, session.CookieName, token, 0)
	return true
}

func (h *Context) ClearAuth(w http.ResponseWriter) {
	h.Cookie.Clear(w, session.CookieName)
}

func (h *Context) RefreshAuth(w http.ResponseWriter, r *http.Request) bool {
	claims, ok := h.Claims(r)
	if !ok {
		return false
	}

	return h.SetAuth(w, claims.PubKey)
}

func (h *Context) EnsureChallengeSession(w http.ResponseWriter, r *http.Request) (string, error) {
	if sessionID := h.Cookie.Value(r, challenge.SessionCookieName); sessionID != "" {
		return sessionID, nil
	}

	sessionID, err := challenge.NewSessionID()
	if err != nil {
		return "", err
	}

	h.Cookie.Set(w, challenge.SessionCookieName, sessionID, 0)
	return sessionID, nil
}

func (h *Context) CurrentOrIssueChallenge(sessionID string) (challenge.Value, error) {
	current, err := h.Challenge.Current(sessionID)
	if err == nil && current.Token != "" {
		return current, nil
	}

	return h.Challenge.Issue(sessionID)
}

func (h *Context) Profile(r *http.Request) *nostr.Profile {
	return h.Account.Read(r)
}

func (h *Context) FetchProfile(ctx context.Context, pubkey string) (*nostr.Profile, error) {
	return h.NostrAccounts.FetchProfile(ctx, pubkey)
}

func (h *Context) ProfileProp(w http.ResponseWriter, r *http.Request, _ string) any {
	pubkey := AuthenticatedPubkeyFromContext(r)
	currentProfile := h.Profile(r)
	if currentProfile == nil && pubkey != "" {
		cachedProfile := h.NostrAccounts.CachedProfile(pubkey)
		if cachedProfile != nil {
			currentProfile = cachedProfile
			h.Account.Set(w, cachedProfile)
		}
	}

	profile := any(currentProfile)
	if pubkey != "" && currentProfile == nil {
		profile = gonertia.Defer(func(ctx context.Context) (any, error) {
			profile, err := h.FetchProfile(ctx, pubkey)
			if err != nil {
				return nil, nil
			}
			if profile == nil {
				return nil, nil
			}

			h.Account.Set(w, profile)
			return profile, nil
		})
	}

	return profile
}

func (h *Context) Allowed(host, pubkey, nip05 string) bool {
	if h.Authz == nil {
		return false
	}

	return h.Authz.Allowed(host, pubkey, nip05)
}

func (h *Context) WithAuthenticatedPubkey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pubkey := h.AuthenticatedPubkey(r)
		if pubkey != "" {
			h.RefreshAuth(w, r)
		}

		ctx := context.WithValue(r.Context(), authenticatedPubkeyKey, pubkey)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Context) RequireCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !h.CSRF.Valid(r, r.Header.Get("X-CSRF-Token")) {
			h.AuthError(w, r, AuthErrorInvalidCSRF)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h *Context) WithChallenge(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionID := h.Cookie.Value(r, challenge.SessionCookieName)
		if sessionID == "" {
			h.AuthError(w, r, AuthErrorMissingChallenge)
			return
		}

		current, err := h.Challenge.Current(sessionID)
		if err != nil {
			h.AuthError(w, r, AuthErrorChallengeUnavailable)
			return
		}

		ctx := context.WithValue(r.Context(), challengeContextKey, ChallengeContext{
			SessionID: sessionID,
			Challenge: current.Token,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AuthenticatedPubkeyFromContext(r *http.Request) string {
	value, _ := r.Context().Value(authenticatedPubkeyKey).(string)
	return value
}

func ChallengeFromContext(r *http.Request) (*ChallengeContext, bool) {
	challengeContext, ok := r.Context().Value(challengeContextKey).(ChallengeContext)
	if !ok || challengeContext.SessionID == "" || challengeContext.Challenge == "" {
		return nil, false
	}

	return &challengeContext, true
}
