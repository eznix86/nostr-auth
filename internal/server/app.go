package server

import "github.com/go-chi/chi/v5"

type App struct {
	Handlers *Context
	Auth     *Auth
	Home     *Home
	Proxy    *Proxy
	router   chi.Router
}

func (a *App) Routes() chi.Router {
	return a.router
}

func (a *App) SetRouter(router chi.Router) {
	a.router = router
}
