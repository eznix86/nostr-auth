package controller

import "net/http"

type contextKey string

const (
	authenticatedPubkeyKey contextKey = "authenticated_pubkey"
	challengeContextKey    contextKey = "challenge_context"
)

type ChallengeContext struct {
	Challenge   string
	IntendedURL string
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
