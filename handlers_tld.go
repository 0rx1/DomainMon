package main

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleTLDs(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	tlds := make(map[string]bool)
	for _, domain := range s.domains {
		tlds[domain.TLD] = true
	}
	s.mu.RUnlock()

	tldList := make([]string, 0, len(tlds))
	for tld := range tlds {
		tldList = append(tldList, tld)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tldList)
}

func (s *Server) handleTLDDomains(w http.ResponseWriter, r *http.Request) {
	// Extract TLD from path
	path := r.URL.Path
	tld := path[len("/api/v1/tlds/"):]

	s.mu.RLock()
	domains := make([]Domain, 0)
	for _, domain := range s.domains {
		if domain.TLD == tld {
			domains = append(domains, domain)
		}
	}
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domains)
}
