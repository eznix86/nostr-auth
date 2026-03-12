package app

import (
	"github.com/eznix86/nostr-auth/internal/account"
	"github.com/eznix86/nostr-auth/internal/authorization"
	"github.com/eznix86/nostr-auth/internal/challenge"
	"github.com/eznix86/nostr-auth/internal/config"
	"github.com/eznix86/nostr-auth/internal/cookie"
	"github.com/eznix86/nostr-auth/internal/csrf"
	"github.com/eznix86/nostr-auth/internal/inertia"
	"github.com/eznix86/nostr-auth/internal/logger"
	"github.com/eznix86/nostr-auth/internal/nostr"
	"github.com/eznix86/nostr-auth/internal/server"
	"github.com/eznix86/nostr-auth/internal/session"
	"github.com/rs/zerolog"
)

func New(cfg config.Config) (*server.App, error) {
	return NewWithLogger(cfg, logger.Default)
}

func NewWithLogger(cfg config.Config, log zerolog.Logger) (*server.App, error) {
	inertiaApp, err := inertia.New(cfg)
	if err != nil {
		return nil, err
	}

	policy, err := authorization.LoadPolicyFile(cfg.AuthConfigFile)
	if err != nil {
		return nil, err
	}

	jar := cookie.NewJar(cfg.CookieDomain, cfg.CookieSecure)
	h := &server.Context{
		Config:        cfg,
		Log:           log,
		Account:       account.NewCookie(jar),
		Authz:         policy,
		Challenge:     challenge.NewStore(cfg.ChallengeTTL),
		Cookie:        jar,
		CSRF:          csrf.NewGuard(jar),
		Inertia:       inertiaApp,
		NostrAccounts: nostr.NewAccounts(server.DefaultRelays, cfg.ProfileFetchTimeout, cfg.ProfileCacheTTL),
		NostrVerify:   nostr.NewVerify(),
		Session:       session.NewSigner(cfg.AppSecret, cfg.SessionTTL),
	}

	app := &server.App{
		Handlers: h,
		Auth:     &server.Auth{H: h},
		Home:     &server.Home{H: h},
		Logout:   &server.LogoutPage{H: h},
		Proxy:    &server.Proxy{H: h},
	}

	router, err := server.NewRouter(app)
	if err != nil {
		return nil, err
	}
	app.SetRouter(router)

	return app, nil
}
