package inertia

import (
	"net/http"

	assets "github.com/eznix86/nostr-auth"
	"github.com/eznix86/nostr-auth/internal/config"
	gonertia "github.com/romsar/gonertia/v2"
)

type Inertia struct {
	vite *gonertia.ViteInstance
}

func New(cfg config.Config, opts ...gonertia.Option) (*Inertia, error) {
	options := []gonertia.Option{}
	options = append(options, opts...)

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

	return &Inertia{vite: vite}, nil
}

func (a *Inertia) Middleware(next http.Handler) http.Handler {
	return a.vite.Middleware(next)
}

func (a *Inertia) Render(w http.ResponseWriter, r *http.Request, component string, props gonertia.Props) error {
	return a.vite.Render(w, r, component, props)
}

func (a *Inertia) ShareProp(key string, val any) {
	a.vite.ShareProp(key, val)
}

func (a *Inertia) ShareTemplateData(key string, val any) {
	a.vite.ShareTemplateData(key, val)
}

func (a *Inertia) Redirect(w http.ResponseWriter, r *http.Request, location string, status int) {
	a.vite.Location(w, r, location, status)
}
