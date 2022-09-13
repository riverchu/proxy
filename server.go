package proxy

import (
	"sync"
	"time"
)

var defaultServer = NewServer()

// serve ...
func serve() {
	for range time.Tick(refreshInterval) {
		defaultServer.Renew(FilterProxyLevel(MEDIUM))
	}
}

// NewServer ...
func NewServer() (server *Server) {
	defer func() { go func() { server.Reload().Unique().JudgeQuality().Filter(FilterProxyLevel(MEDIUM)) }() }()
	return (&Server{sources: ProxySourceList[:]})
}

// Server ...
type Server struct {
	mu sync.RWMutex
	// sources proxy source
	sources []Source
	// proxies all proxies
	proxies ProxyArray

	set map[string]struct{}
}

// Reload reload all proxies
func (s *Server) Reload() *Server {
	proxies := s.getProxies()
	if len(proxies) == 0 {
		return s
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.proxies = proxies

	return s
}

// Unique unique proxies
func (s *Server) Unique() *Server {
	s.mu.RLock()
	set, proxies := s.unique(s.proxies...)
	s.mu.RUnlock()

	s.mu.Lock()
	s.set = set
	s.proxies = proxies
	s.mu.Unlock()

	return s
}

func (s *Server) unique(proxies ...*Proxy) (map[string]struct{}, []*Proxy) {
	set := make(map[string]struct{}, len(proxies))
	result := make([]*Proxy, 0, len(proxies))

	for _, proxy := range proxies {
		proxyStr := proxy.String()
		if _, ok := set[proxyStr]; !ok {
			result = append(result, proxy)
			set[proxyStr] = struct{}{}
		}
	}

	return set, result
}

// Renew equal to Reload + Unique + JudgeQuality + Filter
func (s *Server) Renew(opts ...FilterOption) *Server {
	proxies := s.getProxies()
	if len(proxies) == 0 {
		return s
	}

	set, proxies := s.unique(proxies...)

	proxies.JudgeQuality()

	proxies = s.filter(proxies, opts...)
	if len(proxies) == 0 {
		return s
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.set = set
	s.proxies = proxies

	return s
}

// getProxies get proxy from sources
func (s *Server) getProxies() (proxies ProxyArray) {
	for _, source := range s.sources {
		proxies = append(proxies, source.GetProxies()...)
	}
	return
}

// GetProxy ...
func (s *Server) GetProxy() *Proxy {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.proxies.Pick()
}

// GetProxies ...
func (s *Server) GetProxies(opts ...FilterOption) ProxyArray {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.filter(s.proxies, opts...)
}

// JudgeQuality ...
func (s *Server) JudgeQuality() *Server {
	s.mu.RLock()
	defer s.mu.RUnlock()

	s.proxies.JudgeQuality()

	return s
}

// Filter ...
func (s *Server) Filter(opts ...FilterOption) *Server {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.proxies = s.filter(s.proxies, opts...)

	return s
}

func (s *Server) filter(proxies []*Proxy, opts ...FilterOption) (ps []*Proxy) {
	for _, p := range proxies {
		pass := true
		for _, opt := range opts {
			if !opt(p) {
				pass = false
				break
			}
		}
		if pass {
			ps = append(ps, p)
		}
	}
	return ps
}
