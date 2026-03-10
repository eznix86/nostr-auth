package inertia

import (
	"net/http"

	assets "github.com/eznix86/nostr-auth"
	"github.com/eznix86/nostr-auth/internal/config"
	gonertia "github.com/romsar/gonertia/v2"
)

type App struct {
	vite *gonertia.ViteInstance
}

func New(cfg config.Config) (*App, error) {
	options := []gonertia.Option{}
	if cfg.SSRURL != "" {
		options = append(options, gonertia.WithSSR(cfg.SSRURL))
	}

	root, err := gonertia.NewFromFileFS(assets.AssetsFS, "resources/views/app.html", options...)
	if err != nil {
		return nil, err
	}

	vite, err := gonertia.NewViteFromFS(root, assets.AssetsFS,
		gonertia.WithEntryPoints("resources/js/app.jsx", "resources/css/app.css"),
	)
	if err != nil {
		return nil, err
	}

	return &App{vite: vite}, nil
}

func (a *App) Middleware(next http.Handler) http.Handler {
	return a.vite.Middleware(next)
}

func (a *App) Render(w http.ResponseWriter, r *http.Request, component string, props gonertia.Props) error {
	return a.vite.Render(w, r, component, props)
}

func (a *App) Redirect(w http.ResponseWriter, r *http.Request, location string, status int) {
	a.vite.Location(w, r, location, status)
}
