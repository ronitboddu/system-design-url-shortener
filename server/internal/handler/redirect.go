package handler

import (
	"net/http"
	"strings"

	"www.urlshortener.com/server/internal/util"
)

func (h *Handler) Redirect(rw http.ResponseWriter, req *http.Request) {
	util.CheckGetReq(&rw, req)

	shortCode := strings.TrimPrefix(req.URL.Path, "/")
	urlResponse, err := h.shortenerService.GetRecord(shortCode)

	if err != nil {
		http.NotFound(rw, req)
		return
	}

	http.Redirect(rw, req, urlResponse.OriginalURL, http.StatusFound)
}
