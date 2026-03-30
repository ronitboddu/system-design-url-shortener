package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (h *Handler) TinyUrl(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(rw, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	urlResponse, err := h.shortenerService.PutRecord(req)

	if err != nil {
		http.NotFound(rw, req)
		return
	}

	shortenedUrl := "http://localhost:8080/" + urlResponse.ShortCode

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
