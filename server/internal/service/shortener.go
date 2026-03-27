package service

import (
	"net/http"

	"www.urlshortener.com/server/internal/util"
)

type Process interface {
	GenerateTinyUrl(req *http.Request) (string, string)
}

type Shortener struct {
	ExpTime int    `json:"expTime"`
	UrlPath string `json:"urlPath"`
}

func NewShortner() *Shortener {
	return &Shortener{}
}

func (s *Shortener) GenerateTinyUrl(req *http.Request) (string, string, string) {
	util.DecodeReq(req, s)

	client_ip := util.GetClientIP(req)
	ip_url := client_ip + s.UrlPath
	code := util.GetCode(ip_url)
	shortenedUrl := "http://localhost:8080/" + code
	return code, shortenedUrl, s.UrlPath
}
