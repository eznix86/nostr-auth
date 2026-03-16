package controller

import (
	"context"
	"crypto/subtle"
	"net/http"

	"github.com/eznix86/nostr-auth/internal/session"
)

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
			h.AuthError(w, r, AuthErrorChallengeExpired)
			return
		}

		claims, err := h.Session.VerifyChallenge(signed)
		if err != nil {
			h.AuthError(w, r, AuthErrorChallengeExpired)
			return
		}

		ctx := context.WithValue(r.Context(), challengeContextKey, ChallengeContext{
			Challenge:   claims.Token,
			IntendedURL: claims.IntendedURL,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func subtleConstantTimeCompare(left, right string) bool {
	return subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}
