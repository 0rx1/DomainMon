package main

func (s *Server) updateStats() {
	stats := DomainStats{
		TotalDomains:   len(s.domains),
		DomainsPerTLD:  make(map[string]int),
		LastUpdateTime: s.lastUpdate,
	}

	for _, domain := range s.domains {
		stats.DomainsPerTLD[domain.TLD]++
	}

	s.stats = stats
}
