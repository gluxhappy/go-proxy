package server

import (
	"io"
	"log"
	"net/http"
	"net/url"
)

func (s *ProxyServer) cloneRequest(r *http.Request) *http.Request {
	// Clone the request
	req := &http.Request{}
	req.URL = r.URL
	req.Method = r.Method

	// Clone the request headers
	req.Header = make(http.Header)
	for key, values := range r.Header {
		req.Header[key] = values
	}
	return req
}

func (s *ProxyServer) getCachedHttpClient(proxy string) (*http.Client, bool) {
	s.clientMu.Lock()
	defer s.clientMu.Unlock()
	// Check if there is a proxy for the request
	if cachedClient, ok := s.httpClientsCaches[proxy]; ok {
		return cachedClient, true
	}
	return nil, false
}

func (s *ProxyServer) addHttpClientCache(proxy string) *http.Client {
	var client *http.Client
	s.clientMu.Lock()
	defer s.clientMu.Unlock()
	if proxy == "DIRECT" {
		client = &http.Client{}
	} else {
		proxyURL, _ := url.Parse(proxy)
		client = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		}
	}
	s.httpClientsCaches[proxy] = client
	return client
}
func (s *ProxyServer) handleHttpForward(host string, port string, w http.ResponseWriter, r *http.Request) {
	proxyResult, err := s.findProxy(host)
	if err != nil {
		log.Printf("%s %s %s DENIED/no-mactched rule, %s\n", r.Method, r.Host, r.URL.Path, err)
		http.Error(w, "Error forwarding request, no proxy found", http.StatusForbidden)
		return
	}

	client, cached := s.getCachedHttpClient(proxyResult.proxy)
	if !cached {
		client = s.addHttpClientCache(proxyResult.proxy)
	}

	if proxyResult.proxy == "DENY" {
		log.Printf("%s %s %s DENIED/policy from %s\n", r.Method, r.Host, r.URL.Path, proxyResult.rule)
		http.Error(w, "Error forwarding request, DENIED", http.StatusForbidden)
		return

	} else if proxyResult.proxy == "DIRECT" {
		if !s.config.AllowLocalhost {
			if resolvedHost, err := s.resolver.resolveHost(host); err == nil {
				if resolvedHost.localhost {
					log.Printf("%s %s %s DENIED/localhost\n", r.Method, r.Host, r.URL.Path)
					http.Error(w, "Error forwarding request", http.StatusForbidden)
					return
				}
			} else {
				log.Printf("%s %s %s DENIED/unresolvable, %s\n", r.Method, r.Host, r.URL.Path, err)
				http.Error(w, "Error forwarding request, unresolvable", http.StatusForbidden)
				return
			}
		}
		log.Printf("%s %s %s %s from %s\n", r.Method, r.Host, r.URL.Path, proxyResult.proxy, proxyResult.rule)
	} else {
		log.Printf("%s %s %s %s from %s\n", r.Method, r.Host, r.URL.Path, proxyResult.proxy, proxyResult.rule)
	}

	// Clone the incoming request
	req := s.cloneRequest(r)

	// Perform the request to the destination server
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Error forwarding request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy the response headers to the client
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set the response status code
	w.WriteHeader(resp.StatusCode)

	// Copy the response body to the client
	io.Copy(w, resp.Body)
}
