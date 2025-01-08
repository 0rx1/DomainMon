package main

import (
	"encoding/json"
	"net/http"
	"time"
)

func (s *Server) handleNewDomains(w http.ResponseWriter, r *http.Request) {
	cutoff := time.Now().Add(-24 * time.Hour)

	s.mu.RLock()
	newDomains := make([]Domain, 0)
	for _, domain := range s.domains {
		if domain.CreatedAt.After(cutoff) {
			newDomains = append(newDomains, domain)
		}
	}
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newDomains)
}

func (s *Server) handleRemovedDomains(w http.ResponseWriter, r *http.Request) {
	// This would require tracking removed domains in the Server struct
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]Domain{})
}
