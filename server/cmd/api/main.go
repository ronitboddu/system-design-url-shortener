package main

import (
	"log"
	"net/http"

	"www.urlshortener.com/server/internal/handler"
	"www.urlshortener.com/server/internal/store"
)

func main() {
	mux := http.NewServeMux()

	memstore := store.NewMemoryStore()
	h := handler.NewHandler(memstore)

	mux.HandleFunc("/shorten", h.TinyUrl)
	mux.HandleFunc("/", h.Redirect)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Printf("Starting server on :8080...")
	log.Fatal(srv.ListenAndServe())
}
