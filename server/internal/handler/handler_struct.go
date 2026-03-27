package handler

import (
	"net/http"
	"strings"

	"www.urlshortener.com/server/internal/store"
)

type Handler struct {
	store store.Store
}

func NewHandler(s store.Store) *Handler {
	return &Handler{store: s}
}

func (h *Handler) CheckUrl(urlPath string, rw http.ResponseWriter) {
	if strings.TrimSpace(urlPath) == "" {
		http.Error(rw, "urlPath is required", http.StatusBadRequest)
		return
	}
}
