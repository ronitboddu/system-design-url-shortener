package handler

import (
	"net/http"
	"strings"

	"www.urlshortener.com/server/internal/service"
)

type Handler struct {
	shortenerService *service.Shortener
}

func NewHandler(s *service.Shortener) *Handler {
	return &Handler{shortenerService: s}
}

func (h *Handler) CheckUrl(urlPath string, rw http.ResponseWriter) {
	if strings.TrimSpace(urlPath) == "" {
		http.Error(rw, "urlPath is required", http.StatusBadRequest)
		return
	}
}
