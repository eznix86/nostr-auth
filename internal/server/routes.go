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

	r := chi.NewRouter()
	r.Use(app.Handlers.WithAuthenticatedPubkey)
	r.Get("/healthz", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if !app.IsReady() {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("draining"))
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	r.HandleFunc("/auth/check", app.Proxy.Check)
	r.HandleFunc("/auth/check/*", app.Proxy.Check)
	r.Handle("/public/*", http.StripPrefix("/public/", http.FileServer(http.FS(publicAssets))))
	r.Handle("/build/*", http.StripPrefix("/build/", http.FileServer(http.FS(builtAssets))))

	r.Group(func(r chi.Router) {
		r.Use(app.Handlers.Inertia.Middleware)
		r.Get("/", app.Home.Index)
		r.Get("/auth/csrf", app.Auth.CSRF)
		r.Get("/auth/challenge", app.Auth.Challenge)
		r.Get("/logout", app.Logout.Index)
		r.Group(func(r chi.Router) {
			r.Use(app.Handlers.RequireCSRF)
			r.With(app.Handlers.WithChallenge).Post("/auth/verify", app.Auth.Verify)
			r.Post("/logout", app.Auth.Logout)
		})
	})

	return r, nil
}
