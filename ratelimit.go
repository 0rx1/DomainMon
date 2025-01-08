package main

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type DomainLimiter struct {
	limiters sync.Map // map[string]*rate.Limiter
}

func (dl *DomainLimiter) getLimiter(domain string) *rate.Limiter {
	limiter, exists := dl.limiters.Load(domain)
	if !exists {
		limiter = rate.NewLimiter(rate.Every(time.Second), 2)
		dl.limiters.Store(domain, limiter)
	}
	return limiter.(*rate.Limiter)
}
