package app

import (
	"github.com/eznix86/nostr-auth/internal/account"
	"github.com/eznix86/nostr-auth/internal/appconfig"
	"github.com/eznix86/nostr-auth/internal/authorization"
	"github.com/eznix86/nostr-auth/internal/config"
	"github.com/eznix86/nostr-auth/internal/controller"
	"github.com/eznix86/nostr-auth/internal/cookie"
	"github.com/eznix86/nostr-auth/internal/flash"
	"github.com/eznix86/nostr-auth/internal/inertia"
	"github.com/eznix86/nostr-auth/internal/logger"
	"github.com/eznix86/nostr-auth/internal/nostr"
	"github.com/eznix86/nostr-auth/internal/server"
	"github.com/eznix86/nostr-auth/internal/session"
	gonertia "github.com/romsar/gonertia/v2"
	"github.com/rs/zerolog"
)

func New(cfg config.Config) (*server.App, error) {
	return NewWithLogger(cfg, logger.Default)
}

func NewWithLogger(cfg config.Config, log zerolog.Logger) (*server.App, error) {
	appConfig, err := appconfig.LoadFile(cfg.ConfigFile)
	if err != nil {
		return nil, err
	}

	jar := cookie.NewJar(cfg.CookieDomain, cfg.CookieSecure)

	inertiaApp, err := inertia.New(cfg, gonertia.WithFlashProvider(&flash.Provider{}))
	if err != nil {
		return nil, err
	}

	authorizer, err := authorization.Compile(appConfig.Auth.AuthorizationConfig(), log)
	if err != nil {
		return nil, err
	}

	inertiaApp.ShareProp("branding", appConfig.Branding)
	inertiaApp.ShareTemplateData("backgroundAsset", appConfig.Branding.BackgroundAsset())

	h := &controller.Handler{
		Config:          cfg,
		Log:             log,
		Cookie:          jar,
		Account:         account.NewCookie(jar),
		Authz:           authorizer,
		Inertia:         inertiaApp,
		Nostr:           nostr.NewClient(server.DefaultRelays, cfg.ProfileFetchTimeout, cfg.ProfileCacheTTL),
		Session:         session.NewSigner(cfg.AppSecret, cfg.SessionTTL),
		FlashMiddleware: flash.Middleware(jar),
	}

	app := &server.App{Controller: h, Debug: cfg.Debug}

	router, err := server.NewRouter(app)
	if err != nil {
		return nil, err
	}
	app.Router = router

	return app, nil
}
