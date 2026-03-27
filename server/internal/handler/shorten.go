package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"www.urlshortener.com/server/internal/service"
	"www.urlshortener.com/server/internal/util"
)

func (h *Handler) TinyUrl(rw http.ResponseWriter, req *http.Request) {
	util.CheckPostReq(&rw, req)

	p := service.NewShortner()
	code, shortenedUrl, originalURL := p.GenerateTinyUrl(req)

	h.store.Save(code, originalURL)

	printMap(*h.store.GetUrlMap())
	fmt.Println()

	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(map[string]string{
		"short_url": shortenedUrl,
	})
}

func printMap(urlStore map[string]string) {
	for k, v := range urlStore {
		fmt.Printf("Key: %s, Value: %s\n", k, v)
	}
}
