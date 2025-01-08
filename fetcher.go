package main

import (
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

func (s *Server) backgroundFetch() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for {
		if err := s.fetchData(); err != nil {
			log.Printf("Error fetching data: %v", err)
		}

		<-ticker.C
	}
}

func (s *Server) fetchData() error {
	// Fetch new domains
	newDomains, err := fetchFile("https://codeberg.org/webamon/newly_registered_domains/raw/branch/main/new_domains.txt")
	if err != nil {
		return err
	}

	// Update server state
	s.mu.Lock()
	s.domains = parseDomains(newDomains)
	s.lastUpdate = time.Now()
	s.updateStats()
	s.mu.Unlock()

	return nil
}

func fetchFile(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func parseDomains(data string) []Domain {
	lines := strings.Split(data, "\n")
	domains := make([]Domain, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, ".")
		if len(parts) < 2 {
			continue
		}

		tld := parts[len(parts)-1]
		domains = append(domains, Domain{
			Name:      line,
			TLD:       tld,
			CreatedAt: time.Now(),
			Health:    DomainHealth{}, // Will be updated by health checker
		})
	}

	return domains
}
