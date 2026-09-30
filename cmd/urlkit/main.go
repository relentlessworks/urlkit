package main

import (
	"log"
	"net/http"

	"github.com/relentlessworks/urlkit/internal/api"
	"github.com/relentlessworks/urlkit/internal/config"
)

func main() {
	cfg := config.Load()
	handler := api.NewHandler()
	mux := handler.Routes()

	log.Printf("urlkit listening on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, mux); err != nil {
		log.Fatalf("fatal: %v", err)
	}
}
