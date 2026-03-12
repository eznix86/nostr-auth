package server

import (
	"sync/atomic"

	"github.com/go-chi/chi/v5"
)

type App struct {
	Handlers *Context
	Auth     *Auth
	Home     *Home
	Logout   *LogoutPage
	Proxy    *Proxy
	router   chi.Router
	ready    atomic.Bool
}

func (a *App) Routes() chi.Router {
	return a.router
}

func (a *App) SetRouter(router chi.Router) {
	a.router = router
}

func (a *App) SetReady(ready bool) {
	a.ready.Store(ready)
}

func (a *App) IsReady() bool {
	return a.ready.Load()
}
