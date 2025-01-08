package main

import (
	"fmt"
	"net"
	"net/http"
	"runtime"
	"time"
)

var healthClient = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
	},
}

func (s *Server) checkDomainHealth(domain string) DomainHealth {
	health := DomainHealth{}

	// First check DNS resolution
	_, err := net.LookupHost(domain)
	if err != nil {
		health.Error = fmt.Sprintf("DNS resolution failed: %v", err)
		return health
	}

	// Try HTTPS first
	start := time.Now()
	resp, err := healthClient.Get("https://" + domain)
	if err == nil {
		health.IsOnline = true
		health.Protocol = "https"
		health.StatusCode = resp.StatusCode
		health.ResponseTime = time.Since(start)
		resp.Body.Close()
		return health
	}

	// Fall back to HTTP
	start = time.Now()
	resp, err = healthClient.Get("http://" + domain)
	if err == nil {
		health.IsOnline = true
		health.Protocol = "http"
		health.StatusCode = resp.StatusCode
		health.ResponseTime = time.Since(start)
		resp.Body.Close()
		return health
	}

	health.Error = "Domain is not responding to HTTP(S) requests"
	return health
}

func (s *Server) updateDomainsHealth() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	// Create a worker pool with number of workers based on CPU cores
	wp := NewWorkerPool(runtime.NumCPU() * 2)
	wp.Start(s)

	for {
		s.mu.RLock()
		domains := make([]string, len(s.domains))
		for i, d := range s.domains {
			domains[i] = d.Name
		}
		s.mu.RUnlock()

		// Submit jobs
		go func() {
			for _, domain := range domains {
				wp.jobs <- Job{Domain: domain, Type: "health"}
			}
		}()

		// Collect results
		healthMap := make(map[string]DomainHealth)
		for i := 0; i < len(domains); i++ {
			result := <-wp.results
			if result.Health != nil {
				healthMap[result.Domain] = *result.Health
			}
		}

		// Update domain health
		s.mu.Lock()
		for i, domain := range s.domains {
			if health, ok := healthMap[domain.Name]; ok {
				s.domains[i].Health = health
			}
		}
		s.mu.Unlock()

		<-ticker.C
	}
}
