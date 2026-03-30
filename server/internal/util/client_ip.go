package util

import (
	"net"
	"net/http"
	"strings"
)

func GetClientIP(r *http.Request) string {
	// Check common proxy headers
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		addresses := strings.Split(r.Header.Get(header), ",")
		// Take the first IP in the list
		ip := strings.TrimSpace(addresses[0])
		if ip != "" {
			return ip
		}
	}

	// Fallback to RemoteAddr if no proxy headers are found
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "RemoteAddr: " + r.RemoteAddr
	}
	return "host: " + host
}
