# Go Proxy

An HTTP proxy implementation using Golang, support rules.

### Download

[prebuilds](./releases)

### Configuration

Example configuration can be:
```yaml
# Example config file
host: 127.0.0.1
port: 18888
geo: GeoLite2-Country.mmdb # default: GeoLite2-Country.mmdb, path to GeoLite2-Country.mmdb
allow-localhost: false # default: false, true to allow access to the localhost and loopback addresses
ssl-ports: [443, 8443] # default: [443, 8443], safe ports for HTTP CONNECT
proxies:
  fiddler: http://127.0.0.1:8888
  clash: http://127.0.0.1:33210
  default: DIRECT # DENY, default proxy for unmatched requests
rules:
  - name: "cn"
    type: "DOMAIN-SUFFIX"
    match: ".cn"
    proxy: "DIRECT"  # DENY
  - name: "google"
    type: "WILDCARD"
    match: "*.youtube.com"
    proxy: "fiddler"
  - name: "google"
    type: "REGEXP"
    match: ".*google.*"
    proxy: "clash"
  - name: "external"
    type: "GEOIP"
    match: "!CN"
    proxy: "clash"
```

### Build

Build with docker on Windows
```bat
rem Build for Linux amd64
docker run -it --rm -v %CD%:/code -e GOOS=linux -e GOARCH=amd64 -w /code golang:1.21.5 go build -buildvcs=false -o go-proxy-linux-amd64
rem Build for Linux amd64
docker run -it --rm -v %CD%:/code -e GOOS=linux -e GOARCH=arm64 -w /code golang:1.21.5 go build -buildvcs=false -o go-proxy-linux-arm64
rem Build for Windows
docker run -it --rm -v %CD%:/code -e GOOS=windows -e GOARCH=amd64 -w /code golang:1.21.5 go build -buildvcs=false -o go-proxy-windows-amd64.exe
```

Build with docker on Linux
```sh
# Build for Linux amd64
docker run -it --rm -v $(pwd):/code -e GOOS=linux -e GOARCH=amd64 -w /code golang:1.21.5 go build -buildvcs=false -o go-proxy-linux-amd64
# Build for Linux amd64
docker run -it --rm -v $(pwd):/code -e GOOS=linux -e GOARCH=arm64 -w /code golang:1.21.5 go build -buildvcs=false -o go-proxy-linux-arm64
# Build for Windows
docker run -it --rm -v $(pwd):/code -e GOOS=windows -e GOARCH=amd64 -w /code golang:1.21.5 go build -buildvcs=false -o go-proxy-windows-amd64.exe
```