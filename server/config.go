package server

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Rule struct {
	Name  string
	Type  string
	Match string
	Proxy string
}

type ServerConfig struct {
	Host           string
	Port           int
	Geo            string
	AllowLocalhost bool  `yaml:"allow-localhost"`
	SslPorts       []int `yaml:"ssl-ports"`
	Proxies        map[string]string
	Rules          []Rule
}

const EXAMPLE_CONFIG = `# Example config file
host: 127.0.0.1
port: 18888
geo: GeoLite2-Country.mmdb
allow-localhost: false # default: false
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

`

func LoadConfig(filePath string) (*ServerConfig, error) {
	var configFile string
	isDefault := filePath == ""
	if isDefault {
		arg0 := os.Args[0]
		parentDir := filepath.Dir(arg0)
		configFile = filepath.Join(parentDir, "config.yaml")
	} else {
		configFile = filePath
	}
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		if isDefault {
			// Create the config file
			log.Println("config.yaml not found, creating an example config file at " + configFile)
			err := os.WriteFile("config.yaml", []byte(EXAMPLE_CONFIG), 0644)
			if err != nil {
				log.Fatalf("Failed to create config file: %v", err)
				return nil, err
			}
		} else {
			log.Fatalf("Config file not found: %v", err)
			return nil, err
		}
	}
	// Read the YAML file
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
		return nil, err
	}

	// Unmarshal the YAML data into ServerConfig
	var config ServerConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Failed to unmarshal config file: %v", err)
		return nil, err
	}

	return &config, nil
}
