package service

import (
	"context"
	"net/http"

	"www.urlshortener.com/server/internal/client"
	"www.urlshortener.com/server/internal/util"
)

type Process interface {
	GenerateTinyUrl(req *http.Request) (string, string)
}

type Shortener struct {
	dbClient *client.DBService
	ExpTime  int    `json:"expTime"`
	UrlPath  string `json:"urlPath"`
}

func NewShortner(dbClient *client.DBService) *Shortener {
	return &Shortener{
		dbClient: dbClient,
	}
}

func (s *Shortener) PutRecord(req *http.Request) (*client.URLResponse, error) {
	ctx := context.Background()
	util.DecodeReq(req, s)
	client_ip := util.GetClientIP(req)
	// code := s.GenerateTinyUrl(req)

	payload := client.PutRecordRequest{
		OriginalURL: s.UrlPath,
		IpAddr:      client_ip,
		ExpTime:     s.ExpTime,
	}

	urlResponse, err := s.dbClient.PutRecord(ctx, payload)

	if err != nil {
		return nil, err
	}

	return urlResponse, nil
}

func (s *Shortener) GetRecord(code string) (*client.URLResponse, error) {
	ctx := context.Background()

	urlResponse, err := s.dbClient.GetRecord(ctx, code)

	if err != nil {
		return nil, err
	}

	return urlResponse, nil
}
