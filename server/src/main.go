package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

type url_struct struct {
	ExpTime int    `json:"expTime"`
	UrlPath string `json:"urlPath"`
}

func GetClientIP(r *http.Request) string {
	// Check common proxy headers
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		addresses := strings.Split(r.Header.Get(header), ",")
		// Take the first IP in the list
		ip := strings.TrimSpace(addresses[0])
		if ip != "" {
			return header + " ipAddr: " + ip
		}
	}

	// Fallback to RemoteAddr if no proxy headers are found
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "RemoteAddr: " + r.RemoteAddr
	}
	return "host: " + host
}

func shorten(rw http.ResponseWriter, req *http.Request) {
	start := time.Now()

	decoder := json.NewDecoder(req.Body)
	var u url_struct
	err := decoder.Decode(&u)
	if err != nil {
		panic(err)
	}
	log.Printf("path=%s client_ip=%s expTime=%d duration=%s", req.URL.Path, GetClientIP(req), u.ExpTime, time.Since(start))

	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(map[string]string{
		"short_url": "http://localhost:8080/abc123",
	})

}

func main() {
	http.HandleFunc("/shorten", shorten)
	log.Printf("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
