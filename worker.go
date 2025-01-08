package main

import (
	"sync"
)

type Job struct {
	Domain string
	Type   string // "health", "whois", "dns"
}

type Result struct {
	Domain string
	Health *DomainHealth
	Whois  *WhoisInfo
	DNS    *DNSInfo
	Error  error
}

type WorkerPool struct {
	numWorkers int
	jobs       chan Job
	results    chan Result
	wg         sync.WaitGroup
}

func NewWorkerPool(numWorkers int) *WorkerPool {
	return &WorkerPool{
		numWorkers: numWorkers,
		jobs:       make(chan Job, numWorkers*2),
		results:    make(chan Result, numWorkers*2),
	}
}

func (wp *WorkerPool) Start(s *Server) {
	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(s)
	}
}

func (wp *WorkerPool) worker(s *Server) {
	defer wp.wg.Done()

	for job := range wp.jobs {
		result := Result{Domain: job.Domain}

		switch job.Type {
		case "health":
			health := s.checkDomainHealth(job.Domain)
			result.Health = &health
		case "whois":
			whois := s.performWhoisLookup(job.Domain)
			result.Whois = &whois
		case "dns":
			dns := s.performDNSLookup(job.Domain)
			result.DNS = &dns
		}

		wp.results <- result
	}
}
