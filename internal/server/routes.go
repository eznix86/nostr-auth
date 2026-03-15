package server

import (
	"fmt"
	"io/fs"
	"net/http"

	assets "github.com/eznix86/nostr-auth"
	"github.com/go-chi/chi/v5"
)

func NewRouter(app *App) (chi.Router, error) {
	builtAssets, err := fs.Sub(assets.AssetsFS, "public/build")
	if err != nil {
		return nil, fmt.Errorf("failed to mount built assets: %w", err)
	}

	publicAssets, err := fs.Sub(assets.AssetsFS, "public")
	if err != nil {
		return nil, fmt.Errorf("failed to mount public assets: %w", err)
	}

	h := app.Handler

	r := chi.NewRouter()
	r.Use(h.WithAuthenticatedPubkey)
	r.Get("/healthz", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if !app.IsReady() {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("draining"))
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	r.HandleFunc("/auth/check", h.ProxyCheck)
	r.HandleFunc("/auth/check/*", h.ProxyCheck)
	r.Handle("/public/*", http.StripPrefix("/public/", http.FileServer(http.FS(publicAssets))))
	r.Handle("/build/*", http.StripPrefix("/build/", http.FileServer(http.FS(builtAssets))))

	r.Group(func(r chi.Router) {
		r.Use(h.FlashMiddleware)
		r.Use(h.Inertia.Middleware)
		r.Get("/", h.HomeIndex)
		r.Get("/auth/csrf", h.AuthCSRF)
		r.Get("/auth/challenge", h.AuthChallenge)
		r.Get("/logout", h.LogoutIndex)
		r.Group(func(r chi.Router) {
			r.Use(h.RequireCSRF)
			r.With(h.WithChallenge).Post("/auth/verify", h.AuthVerify)
			r.Post("/logout", h.LogoutSubmit)
		})
	})

	return r, nil
}
