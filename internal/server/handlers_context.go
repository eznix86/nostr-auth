package server

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"

	"github.com/eznix86/nostr-auth/internal/account"
	"github.com/eznix86/nostr-auth/internal/authorization"
	"github.com/eznix86/nostr-auth/internal/config"
	"github.com/eznix86/nostr-auth/internal/cookie"
	"github.com/eznix86/nostr-auth/internal/inertia"
	"github.com/eznix86/nostr-auth/internal/nostr"
	"github.com/eznix86/nostr-auth/internal/session"
	gonertia "github.com/romsar/gonertia/v2"
	"github.com/rs/zerolog"
)

type contextKey string

const (
	authenticatedPubkeyKey contextKey = "authenticated_pubkey"
	challengeContextKey    contextKey = "challenge_context"

	CSRFCookieName = "nostr_auth_csrf"
)

type ChallengeContext struct {
	Challenge   string
	IntendedURL string
}

type Handler struct {
	Config          config.Config
	Log             zerolog.Logger
	Cookie          *cookie.Jar
	Account         *account.Cookie
	Authz           *authorization.Policy
	Inertia         *inertia.Inertia
	Nostr           *nostr.Client
	Session         *session.Signer
	FlashMiddleware func(http.Handler) http.Handler
}

func (h *Handler) Text(w http.ResponseWriter, status int, body string) {
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

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

func (h *Handler) AuthenticatedPubkey(r *http.Request) string {
	claims, ok := h.Claims(r)
	if !ok {
		return ""
	}

	return claims.PubKey
}

func (h *Handler) Claims(r *http.Request) (*session.Claims, bool) {
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

func (h *Handler) SetAuth(w http.ResponseWriter, pubkey string) bool {
	token, err := h.Session.Sign(pubkey)
	if err != nil {
		return false
	}

	h.Cookie.Set(w, session.CookieName, token, 0)
	return true
}

func (h *Handler) ClearAuth(w http.ResponseWriter) {
	h.Cookie.Clear(w, session.CookieName)
}

func (h *Handler) RefreshAuth(w http.ResponseWriter, r *http.Request) bool {
	claims, ok := h.Claims(r)
	if !ok {
		return false
	}

	return h.SetAuth(w, claims.PubKey)
}

func (h *Handler) SetIntendedURL(w http.ResponseWriter, intendedURL string) bool {
	if intendedURL == "" {
		h.Cookie.Clear(w, session.IntendedURLCookieName)
		return true
	}

	token, err := h.Session.SignIntendedURL(intendedURL, h.Config.ChallengeTTL)
	if err != nil {
		return false
	}

	h.Cookie.Set(w, session.IntendedURLCookieName, token, 0)
	return true
}

func (h *Handler) IntendedURL(r *http.Request) string {
	token := h.Cookie.Value(r, session.IntendedURLCookieName)
	if token == "" {
		return ""
	}

	intendedURL, err := h.Session.VerifyIntendedURL(token)
	if err != nil {
		return ""
	}

	return intendedURL
}

func (h *Handler) ClearIntendedURL(w http.ResponseWriter) {
	h.Cookie.Clear(w, session.IntendedURLCookieName)
}

func (h *Handler) IssueChallenge(w http.ResponseWriter, intendedURL string) (string, error) {
	token, err := randomHex(16)
	if err != nil {
		return "", err
	}

	signed, err := h.Session.SignChallenge(token, intendedURL, h.Config.ChallengeTTL)
	if err != nil {
		return "", err
	}

	h.Cookie.Set(w, session.ChallengeCookieName, signed, 0)
	return token, nil
}

func (h *Handler) Profile(r *http.Request) *nostr.Profile {
	return h.Account.Read(r)
}

func (h *Handler) FetchProfile(ctx context.Context, pubkey string) (*nostr.Profile, error) {
	return h.Nostr.FetchProfile(ctx, pubkey)
}

func (h *Handler) ProfileProp(w http.ResponseWriter, r *http.Request, _ string) any {
	pubkey := AuthenticatedPubkeyFromContext(r)
	currentProfile := h.Profile(r)
	if currentProfile == nil && pubkey != "" {
		cachedProfile := h.Nostr.CachedProfile(pubkey)
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

func (h *Handler) Allowed(host, pubkey, nip05 string) bool {
	if h.Authz == nil {
		return false
	}

	return h.Authz.Allowed(host, pubkey, nip05)
}

func (h *Handler) Groups(host, pubkey, nip05 string) []string {
	if h.Authz == nil {
		return nil
	}

	return h.Authz.Groups(host, pubkey, nip05)
}

func (h *Handler) ensureCSRF(w http.ResponseWriter, r *http.Request) (string, error) {
	if token := h.Cookie.Value(r, CSRFCookieName); token != "" {
		return token, nil
	}

	token, err := randomHex(16)
	if err != nil {
		return "", err
	}

	h.Cookie.Set(w, CSRFCookieName, token, 0)
	return token, nil
}

func (h *Handler) validCSRF(r *http.Request, token string) bool {
	if token == "" {
		return false
	}

	cookieValue := h.Cookie.Value(r, CSRFCookieName)
	if cookieValue == "" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(cookieValue), []byte(token)) == 1
}

func (h *Handler) clearCSRF(w http.ResponseWriter) {
	h.Cookie.Clear(w, CSRFCookieName)
}

func (h *Handler) WithAuthenticatedPubkey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pubkey := h.AuthenticatedPubkey(r)
		if pubkey != "" {
			h.RefreshAuth(w, r)
		}

		ctx := context.WithValue(r.Context(), authenticatedPubkeyKey, pubkey)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) RequireCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !h.validCSRF(r, r.Header.Get("X-CSRF-Token")) {
			h.AuthError(w, r, AuthErrorInvalidCSRF)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h *Handler) WithChallenge(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		signed := h.Cookie.Value(r, session.ChallengeCookieName)
		if signed == "" {
			h.AuthError(w, r, AuthErrorMissingChallenge)
			return
		}

		claims, err := h.Session.VerifyChallenge(signed)
		if err != nil {
			h.AuthError(w, r, AuthErrorChallengeUnavailable)
			return
		}

		ctx := context.WithValue(r.Context(), challengeContextKey, ChallengeContext{
			Challenge:   claims.Token,
			IntendedURL: claims.IntendedURL,
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
	if !ok || challengeContext.Challenge == "" {
		return nil, false
	}

	return &challengeContext, true
}

func randomHex(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}
