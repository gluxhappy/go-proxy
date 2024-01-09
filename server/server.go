package server

import (
	"net/http"
	"sync"
)

type ProxyServer struct {
	config            *ServerConfig
	proxyCaches       map[string]*ProxyResult
	httpClientsCaches map[string]*http.Client
	resolver          *Resolver
	clientMu          sync.Mutex
	cacheMu           sync.Mutex
}

func MakeProxyServer(config *ServerConfig) *ProxyServer {
	return &ProxyServer{config, make(map[string]*ProxyResult), make(map[string]*http.Client), MakeResolver(config.Geo), sync.Mutex{}, sync.Mutex{}}
}

func (s *ProxyServer) HandleRequest(w http.ResponseWriter, r *http.Request) {
	host, port := extractHostPort(r)
	if r.Method == "CONNECT" {
		s.handleHttpConnect(host, port, w, r)
	} else {
		s.handleHttpForward(host, port, w, r)
	}
}
