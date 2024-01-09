# Go Proxy

An HTTP proxy implementation using Golang, support rules.


### Configuration

Example configuration can be:
```yaml
# Example config file
host: 127.0.0.1
port: 18888
geo: GeoLite2-Country.mmdb
proxies:
  fiddler: http://127.0.0.1:8888
  clash: http://127.0.0.1:33210
  default: DIRECT # DENY
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
rem Build for Linux
docker run -it --rm -v %CD%:/code -e GOOS=linux -w /code golang:1.21.5 go build -buildvcs=false
rem Build for Windows
docker run -it --rm -v %CD%:/code -e GOOS=windows -w /code golang:1.21.5 go build -buildvcs=false
```

Build with docker on Linux
```sh
# Build for Linux
docker run -it --rm -v $(pwd):/code -e GOOS=linux -w /code golang:1.21.5 go build -buildvcs=false
# Build for Windows
docker run -it --rm -v $(pwd):/code -e GOOS=windows -w /code golang:1.21.5 go build -buildvcs=false
```