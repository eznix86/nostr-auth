package main

import (
	"log"
	"net/http"

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

	log.Printf("Listening on http://%s", cfg.Address())
	log.Fatal(http.ListenAndServe(cfg.Address(), serverApp.Routes()))
}
