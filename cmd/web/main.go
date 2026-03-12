package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eznix86/nostr-auth/internal/app"
	"github.com/eznix86/nostr-auth/internal/config"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found, using system environment")
	}

	cfg := config.Load()

	serverApp, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	serverApp.SetReady(true)

	httpServer := &http.Server{
		Addr:    cfg.Address(),
		Handler: serverApp.Routes(),
	}

	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)

		sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		<-sigCtx.Done()
		serverApp.SetReady(false)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
		}
	}()

	log.Printf("Listening on http://%s", cfg.Address())
	err = httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}

	<-shutdownDone
}
