package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) handleDomains(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	limit, _ := strconv.Atoi(query.Get("limit"))
	search := query.Get("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	s.mu.RLock()
	filtered := make([]Domain, 0)
	for _, domain := range s.domains {
		if search != "" && !strings.Contains(domain.Name, search) {
			continue
		}
		filtered = append(filtered, domain)
	}
	s.mu.RUnlock()

	start := (page - 1) * limit
	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	response := map[string]interface{}{
		"domains": filtered[start:end],
		"total":   len(filtered),
		"page":    page,
		"limit":   limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	stats := s.stats
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleDomainHealth(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		http.Error(w, "domain parameter is required", http.StatusBadRequest)
		return
	}

	health := s.checkDomainHealth(domain)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}
