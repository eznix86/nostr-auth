package controller

import (
	"net/http"

	"github.com/eznix86/nostr-auth/internal/account"
	"github.com/eznix86/nostr-auth/internal/authorization"
	"github.com/eznix86/nostr-auth/internal/config"
	"github.com/eznix86/nostr-auth/internal/cookie"
	"github.com/eznix86/nostr-auth/internal/inertia"
	"github.com/eznix86/nostr-auth/internal/nostr"
	"github.com/eznix86/nostr-auth/internal/session"
	"github.com/rs/zerolog"
)

type Handler struct {
	Config          config.Config
	Log             zerolog.Logger
	Cookie          *cookie.Jar
	Account         *account.Cookie
	Authz           *authorization.Authorizer
	Inertia         *inertia.Inertia
	Nostr           *nostr.Client
	Session         *session.Signer
	FlashMiddleware func(http.Handler) http.Handler
}
