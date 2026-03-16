package server

import (
	"sync/atomic"

	"github.com/eznix86/nostr-auth/internal/controller"
	"github.com/go-chi/chi/v5"
)

type App struct {
	Controller *controller.Handler
	Router     chi.Router
	Debug      bool
	ready      atomic.Bool
}

func (a *App) SetReady(ready bool) {
	a.ready.Store(ready)
}

func (a *App) IsReady() bool {
	return a.ready.Load()
}
