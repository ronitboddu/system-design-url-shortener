package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"www.urlshortener.com/server/internal/util"
)

func (h *Handler) TinyUrl(rw http.ResponseWriter, req *http.Request) {
	util.CheckPostReq(&rw, req)

	util.DecodeReq(req, h.shortenerService)

	urlResponse, err := h.shortenerService.PutRecord(req)

	if err != nil {
		http.NotFound(rw, req)
		return
	}

	shortenedUrl := "http://localhost:8080" + urlResponse.ShortCode

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
