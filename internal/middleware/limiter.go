package middleware

import (
	"sync"
	"time"
)

type Limiter struct {
	mu   sync.Mutex
	hits map[string][]time.Time
}

func NewLimiter() *Limiter {
	return &Limiter{hits: make(map[string][]time.Time)}
}

func (l *Limiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	a := l.hits[ip]
	cut := now.Add(-time.Minute)
	j := 0
	for _, t := range a {
		if t.After(cut) {
			a[j] = t
			j++
		}
	}
	a = a[:j]
	if len(a) >= 5 {
		l.hits[ip] = a
		return false
	}
	l.hits[ip] = append(a, now)
	return true
}
