# Go Proxy

An HTTP proxy implementation using Golang, support rules.

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