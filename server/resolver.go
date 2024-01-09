package server

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/oschwald/geoip2-golang"
)

type ResolverCache struct {
	ips        []net.IP
	localhost  bool
	country    string
	lastUpdate time.Time
	lastRead   time.Time
}

type ResolvedHost struct {
	localhost bool
	country   string
}

type Resolver struct {
	cache map[string]*ResolverCache
	geoip *geoip2.Reader
	lock  sync.RWMutex
}

func MakeResolver(geodb string) *Resolver {
	if len(geodb) == 0 {
		geodb = "GeoLite2-Country.mmdb"
		log.Println("Using default GeoIP database:" + geodb)
	}
	db, err := geoip2.Open("GeoLite2-Country.mmdb")
	if err != nil {
		log.Println(err)
	} else {
		log.Println("GeoIP database loaded :" + geodb)
	}
	return &Resolver{make(map[string]*ResolverCache), db, sync.RWMutex{}}
}

func (r *Resolver) getGeoCountry(ip net.IP) (string, error) {
	if r.geoip == nil {
		return "", fmt.Errorf("no geoip database available")
	}
	contry, err := r.geoip.Country(ip)
	if err != nil {
		return "", err
	}
	return contry.Country.IsoCode, nil
}

func (r *Resolver) resolveHost(host string) (result *ResolvedHost, err error) {
	if host == "localhost" || host == "127.0.0.1" {
		return &ResolvedHost{true, "LOCALHOST"}, nil
	}

	cachedValue, ok := r.getCache(host)
	// r.cleanCache()
	if ok {
		cachedValue.lastRead = time.Now()
		return &ResolvedHost{cachedValue.localhost, cachedValue.country}, nil
	} else {
		cachedValue, ok = r.updateCache(host)
		if ok {
			return &ResolvedHost{cachedValue.localhost, cachedValue.country}, nil
		}
	}
	return nil, errors.New("Failed to resolve host " + host)
}

func (r *Resolver) updateCache(host string) (cache *ResolverCache, ok bool) {
	ips, err := net.LookupIP(host)
	var localhost = false
	var geo = ""
	if err != nil {
		log.Println("Failed to resolve "+host, err)
		return nil, false
	}
	for _, ip := range ips {
		if !localhost && ip.IsLoopback() {
			localhost = true
		}
		newGeo, err := r.getGeoCountry(ip)
		if err != nil {
			log.Println("Failed to get geo country for "+ip.String(), err)
			return nil, false
		} else {
			if geo == "" {
				geo = newGeo
			} else if geo != newGeo {
				return nil, false
			}
		}
	}

	r.lock.Lock()
	defer r.lock.Unlock()
	cache = &ResolverCache{ips, localhost, geo, time.Now(), time.Now()}
	r.cache[host] = cache
	return cache, true
}

func (r *Resolver) getCache(host string) (*ResolverCache, bool) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if cachedValue, ok := r.cache[host]; ok {
		return cachedValue, true
	}
	return nil, false
}

func (r *Resolver) CleanCache() {
	if r.lock.TryLock() {
		defer r.lock.Unlock()
	}
	for host, cache := range r.cache {
		if time.Since(cache.lastRead) > 30*time.Minute {
			delete(r.cache, host)
		}
	}
}
