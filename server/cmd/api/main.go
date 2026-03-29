package main

import (
	"log"
	"net/http"

	"www.urlshortener.com/server/internal/client"
	"www.urlshortener.com/server/internal/config"
	"www.urlshortener.com/server/internal/handler"
	"www.urlshortener.com/server/internal/service"
)

func main() {
	mux := http.NewServeMux()

	// memstore := store.NewMemoryStore()
	cfg := config.Load()
	dbClient := client.NewDBService(cfg.DBServiceBaseURL)
	shortenerService := service.NewShortner(dbClient)
	h := handler.NewHandler(shortenerService)

	mux.HandleFunc("/shorten", h.TinyUrl)
	mux.HandleFunc("/", h.Redirect)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Printf("Starting server on :8080...")
	log.Fatal(srv.ListenAndServe())
}
