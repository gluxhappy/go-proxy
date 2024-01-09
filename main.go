package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"example.com/go-proxy/server"
)

func main() {
	configFile := flag.String("config", "", "Path of the configuration file,default is the config.yaml in the same directory of the executable.")
	flag.Parse()
	config, err := server.LoadConfig(*configFile)
	if err != nil {
		log.Fatal("Failed to load config:", err)
		return
	}
	hostPort := config.Host + ":" + fmt.Sprint(config.Port)
	// Start the proxy server
	log.Println("Proxy server listening on " + hostPort)
	proxyServer := server.MakeProxyServer(config)

	err = http.ListenAndServe(hostPort, http.HandlerFunc(proxyServer.HandleRequest))
	if err != nil {
		log.Fatal("Proxy server error:", err)
	}
}
