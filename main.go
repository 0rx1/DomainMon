package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	domains         []Domain
	stats           DomainStats
	mu              sync.RWMutex
	lastUpdate      time.Time
	cache           *Cache
	workers         *WorkerPool
	similarityCache map[string]*CachedData
	similarityMu    sync.RWMutex
}

type CachedData struct {
	Data      []SimilarityData
	Timestamp time.Time
}

func main() {
	server := &Server{
		cache:           NewCache(15 * time.Minute),
		workers:         NewWorkerPool(runtime.NumCPU() * 2),
		similarityCache: make(map[string]*CachedData),
	}

	server.workers.Start(server)

	// Initialize data fetcher and health checker
	go server.backgroundFetch()
	go server.updateDomainsHealth()

	// Initialize chi router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(server.cors)
	r.Use(server.rateLimiter)
	r.Use(server.compress)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/domains", server.handleDomains)
		r.Get("/domains/new", server.handleNewDomains)
		r.Get("/domains/removed", server.handleRemovedDomains)
		r.Get("/domains/stats", server.handleStats)
		r.Get("/domains/health", server.handleDomainHealth)
		r.Get("/tlds", server.handleTLDs)
		r.Get("/tlds/{tld}", server.handleTLDDomains)
		r.Get("/lookup/whois", server.handleWhoisLookup)
		r.Get("/lookup/dns", server.handleReverseDNS)

		// Add similarity endpoint
		r.Get("/similarity/{threshold}", server.handleSimilarity)
	})

	// Serve static files
	workDir, _ := os.Getwd()
	filesDir := http.Dir(workDir)
	FileServer(r, "/", filesDir)

	log.Fatal(http.ListenAndServe(":8080", r))
}

// FileServer is a convenience function for serving static files
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}

type SimilarityData struct {
	TargetDomain string   `json:"targetDomain"`
	Count        int      `json:"count"`
	Examples     []string `json:"examples"`
	Similarity   float64  `json:"similarity"`
}

type SimilarityResponse struct {
	Data []SimilarityData `json:"data"`
}

func (s *Server) handleSimilarity(w http.ResponseWriter, r *http.Request) {
	threshold := chi.URLParam(r, "threshold")
	if threshold == "" {
		http.Error(w, "threshold parameter is required", http.StatusBadRequest)
		return
	}

	log.Printf("Fetching similarity data for threshold: %s", threshold)

	data, err := s.fetchAndCacheSimilarityData(threshold)
	if err != nil {
		log.Printf("Error fetching similarity data: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600")

	response := SimilarityResponse{Data: data}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *Server) fetchAndCacheSimilarityData(threshold string) ([]SimilarityData, error) {
	s.similarityMu.RLock()
	if cached, ok := s.similarityCache[threshold]; ok {
		if time.Since(cached.Timestamp) < time.Hour {
			s.similarityMu.RUnlock()
			return cached.Data, nil
		}
	}
	s.similarityMu.RUnlock()

	// Create a custom HTTP client with timeout and proper headers
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: true,
		},
	}

	// Update URL to use the new f500_domains path
	url := fmt.Sprintf("https://codeberg.org/webamon/newly_registered_domains/raw/branch/main/monitoring_output/f500_domains/similarity_%s.txt", threshold)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Add necessary headers
	req.Header.Set("User-Agent", "DomainSentinel/1.0")
	req.Header.Set("Accept", "text/plain")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch similarity data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	similarities := make(map[string]*SimilarityData)

	for scanner.Scan() {
		line := scanner.Text()
		if match := regexp.MustCompile(`(.*?) -> (.*?) \(([\d.]+)%\)`).FindStringSubmatch(line); match != nil {
			newDomain := match[1]
			targetDomain := match[2]
			similarity, _ := strconv.ParseFloat(match[3], 64)

			if _, exists := similarities[targetDomain]; !exists {
				similarities[targetDomain] = &SimilarityData{
					TargetDomain: targetDomain,
					Count:        0,
					Examples:     make([]string, 0, 3),
					Similarity:   similarity,
				}
			}

			similarities[targetDomain].Count++
			if len(similarities[targetDomain].Examples) < 3 {
				similarities[targetDomain].Examples = append(similarities[targetDomain].Examples, newDomain)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	if len(similarities) == 0 {
		return nil, fmt.Errorf("no similarity data found")
	}

	result := make([]SimilarityData, 0, len(similarities))
	for _, data := range similarities {
		result = append(result, *data)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})

	if len(result) > 10 {
		result = result[:10]
	}

	s.similarityMu.Lock()
	s.similarityCache[threshold] = &CachedData{
		Data:      result,
		Timestamp: time.Now(),
	}
	s.similarityMu.Unlock()

	return result, nil
}
