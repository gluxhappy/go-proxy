package server

import (
	"errors"
	"log"
	"regexp"
	"strings"
)

type RuleProcessor struct {
	separator   string
	processFunc func(resolver *Resolver, rule string, host string) (bool, error)
}

type ProxyResult struct {
	proxy string
	rule  string
}

var ruleProcessors = make(map[string]*RuleProcessor)

func init() {
	ruleProcessors["DOMAIN-SUFFIX"] = &RuleProcessor{";",
		func(resolver *Resolver, rule, host string) (bool, error) {
			return strings.HasSuffix(host, rule), nil
		}}
	ruleProcessors["WILDCARD"] = &RuleProcessor{";",
		func(resolver *Resolver, rule, host string) (bool, error) {
			return matchWilecard(rule, host), nil
		}}
	ruleProcessors["REGEXP"] = &RuleProcessor{";",
		func(resolver *Resolver, rule string, host string) (bool, error) {
			match, err := regexp.MatchString(rule, host)
			if err != nil {
				log.Println("Error matching regular expression:", rule, err)
				return false, err
			}
			return match, nil
		}}
	ruleProcessors["GEOIP"] = &RuleProcessor{";",
		func(resolver *Resolver, rule string, host string) (bool, error) {
			resolved, err := resolver.resolveHost(host)
			if err != nil {
				return false, err
			}

			return resolved.country == rule, nil
		}}
}

func (s *ProxyServer) tryGetProxy(host string) (*ProxyResult, bool) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	// Check if there is a proxy for the request
	if cachedProxy, ok := s.proxyCaches[host]; ok {
		return cachedProxy, true
	}
	return nil, false
}

func (s *ProxyServer) findProxy(host string) (*ProxyResult, error) {
	if proxy, ok := s.tryGetProxy(host); ok {
		return proxy, nil
	}
	finalProxy, err := s.processRules(host)
	if err != nil {
		finalProxy = &ProxyResult{s.getProxy("default"), "default"}
	}
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	s.proxyCaches[host] = finalProxy
	return finalProxy, nil
}

func (s *ProxyServer) processRules(host string) (*ProxyResult, error) {
	for _, rule := range s.config.Rules {
		ruleProcessor, ok := ruleProcessors[rule.Type]
		if !ok {
			continue
		}
		ruleMatchItem := []string{rule.Match}
		if len(ruleProcessor.separator) > 0 && strings.Contains(rule.Match, ruleProcessor.separator) {
			ruleMatchItem = strings.Split(rule.Match, ruleProcessor.separator)
		}
		for _, ruleMatch := range ruleMatchItem {
			inverted := false
			if strings.HasPrefix(ruleMatch, "!") {
				inverted = true
				ruleMatch = ruleMatch[1:]
			}

			matched, err := ruleProcessor.processFunc(s.resolver, ruleMatch, host)
			if err != nil {
				log.Printf("Processing rule failed: %s, %s, %s", rule.Type, host, err)
				continue
			}
			if (!inverted && matched) || (inverted && !matched) {
				return &ProxyResult{s.getProxy(rule.Proxy), rule.Name}, nil
			}
		}
	}
	return nil, errors.New("no rule match found for " + host)
}

func (s *ProxyServer) getProxy(target string) string {
	if s.config.Proxies == nil {
		return "DIRECT"
	}
	if target == "DIRECT" || target == "DENY" {
		return target
	}
	if proxy, ok := s.config.Proxies[target]; ok {
		return proxy
	}
	if proxy, ok := s.config.Proxies["default"]; ok {
		return proxy
	}
	return "DIRECT"
}
