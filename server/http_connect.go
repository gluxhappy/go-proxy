package server

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

func (s *ProxyServer) handleHttpConnect(host string, port string, w http.ResponseWriter, r *http.Request) {
	intPort, err := strconv.Atoi(port)
	if err != nil {
		log.Println("CONNECT: " + r.URL.Host + " DENIED/ilegal-port:" + port)
		http.Error(w, "The given port is not secure, DENY: "+r.Host, http.StatusForbidden)
		return
	}
	safePort := false
	for _, sslPort := range s.config.SslPorts {
		if sslPort == intPort {
			safePort = true
			break
		}
	}

	if !safePort {
		log.Println("CONNECT: " + r.URL.Host + " DENIED/non-ssl-port:" + port)
		http.Error(w, "The given port is not secure, DENY: "+r.Host, http.StatusForbidden)
		return
	}

	proxyResult, err := s.findProxy(host)
	if err != nil {
		log.Println("CONNECT: "+r.URL.Host+" DENIED/no-mactched rule", err)
		http.Error(w, "No proxy found for the given host, DENY: "+r.Host, http.StatusForbidden)
		return
	}
	if proxyResult.proxy == "DENY" {
		log.Println("CONNECT: " + r.URL.Host + " DENIED/policy from " + proxyResult.rule)
		http.Error(w, "The given host is a denied, DENY: "+r.Host, http.StatusForbidden)
		return
	} else if proxyResult.proxy == "DIRECT" {
		if !s.config.AllowLocalhost {
			if resolveHost, err := s.resolver.resolveHost(host); err == nil {
				if resolveHost.localhost {
					log.Println("CONNECT: " + r.URL.Host + " DENIED/localhost")
					http.Error(w, "The given host is a localhost, DENY: "+r.Host, http.StatusForbidden)
					return
				}
			} else {
				log.Println("CONNECT: "+r.URL.Host+" DENIED/unresolvable", err)
				http.Error(w, "The given host is a unresolvable, DENY: "+r.Host, http.StatusForbidden)
				return
			}
		}
		handleDirectConnect(proxyResult, host, port, w, r)
	} else {
		handleUpStreamProxy(proxyResult, w, r)
	}
}

func handleUpStreamProxy(proxyResult *ProxyResult, w http.ResponseWriter, r *http.Request) {
	// Create a new HTTP client with a proxy
	log.Println("CONNECT: " + r.URL.Host + " " + proxyResult.proxy + " from " + proxyResult.rule)
	proxyURL, _ := url.Parse(proxyResult.proxy)
	proxyConn, err := net.Dial("tcp", proxyURL.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer proxyConn.Close()
	proxyConn.Write([]byte("CONNECT " + r.URL.Host + " HTTP/1.1\r\n\r\n"))
	br := bufio.NewReader(proxyConn)
	resp, err := http.ReadResponse(br, nil)
	if err != nil {
		log.Println("buffer: ", err)
		return
	}
	if resp.StatusCode != 200 {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	// clientConn.Write([]byte("\r\n"))
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go transfer(proxyConn, clientConn, wg)
	go transfer(clientConn, proxyConn, wg)
	wg.Wait()
}

func handleDirectConnect(proxyResult *ProxyResult, host string, port string, w http.ResponseWriter, r *http.Request) {
	log.Println("CONNECT: " + r.URL.Host + " DIRECT from " + proxyResult.rule)
	// Connect to the destination server
	destConn, err := net.Dial("tcp", host+":"+port)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer destConn.Close()

	// Let the client know the connection was established
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	// clientConn.Write([]byte("\r\n"))
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go transfer(destConn, clientConn, wg)
	go transfer(clientConn, destConn, wg)
	wg.Wait()
}

// Copy the client's connection to the destination and vice versa
func transfer(src io.ReadCloser, dest io.WriteCloser, wg *sync.WaitGroup) {
	// Copy data from source to destination
	defer wg.Done()
	_, err := io.Copy(dest, src)
	if err != nil && err != io.EOF {
		if _, ok := src.(*net.TCPConn); !ok {
			log.Println("Error transferring data:", err)
		}
	}
}
