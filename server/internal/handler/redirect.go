package handler

import (
	"net/http"
	"strings"

	"www.urlshortener.com/server/internal/util"
)

func (h *Handler) Redirect(rw http.ResponseWriter, req *http.Request) {
	util.CheckGetReq(&rw, req)

	shortCode := strings.TrimPrefix(req.URL.Path, "/")
	originalURL, ok := h.store.Get(shortCode)

	if !ok {
		http.NotFound(rw, req)
		return
	}

	http.Redirect(rw, req, originalURL, http.StatusFound)
}
