package server

import (
	"net/http"
	"strings"
)

func extractHostPort(r *http.Request) (host string, port string) {
	if r.Method == "CONNECT" {
		host = r.URL.Hostname()
		port = r.URL.Port()
	} else {
		host, port = splitHostPort(r.Host)
		if port == "" {
			if r.URL.Scheme == "https" {
				port = "443"
			} else {
				port = "80"
			}
		}
	}
	return host, port
}

func validOptionalPort(port string) bool {
	if port == "" {
		return true
	}
	if port[0] != ':' {
		return false
	}
	for _, b := range port[1:] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

func splitHostPort(hostPort string) (host, port string) {
	host = hostPort

	colon := strings.LastIndexByte(host, ':')
	if colon != -1 && validOptionalPort(host[colon:]) {
		host, port = host[:colon], host[colon+1:]
	}

	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}

	return
}
