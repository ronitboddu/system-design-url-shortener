package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func (h *Handler) Redirect(rw http.ResponseWriter, req *http.Request) {
	start := time.Now()
	if req.Method != http.MethodGet {
		http.Error(rw, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	shortCode := strings.TrimPrefix(req.URL.Path, "/")
	urlResponse, err := h.shortenerService.GetRecord(shortCode)

	if err != nil {
		http.NotFound(rw, req)
		return
	}
	if urlResponse == nil {
		http.NotFound(rw, req)
		return
	}

	http.Redirect(rw, req, urlResponse.OriginalURL, http.StatusFound)
	elapsed := time.Since(start)
	fmt.Printf("Execution took for Redirect %s\n", elapsed)
}
