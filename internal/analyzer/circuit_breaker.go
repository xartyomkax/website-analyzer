package analyzer

import (
	"sync"
	"time"
)

// circuitBreaker manages failure counts per domain with half-open state support
type circuitBreaker struct {
	mu               sync.RWMutex
	failures         map[string]int
	successes        map[string]int
	lastAttempt      map[string]time.Time
	maxFailures      int
	successThreshold int
	retryDelay       time.Duration
}

func newCircuitBreaker(maxFailures int) *circuitBreaker {
	return &circuitBreaker{
		failures:         make(map[string]int),
		successes:        make(map[string]int),
		lastAttempt:      make(map[string]time.Time),
		maxFailures:      maxFailures,
		successThreshold: 3,
		retryDelay:       2 * time.Second,
	}
}

func (cb *circuitBreaker) allow(domain string) bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	failCount := cb.failures[domain]

	// If not in open state, allow
	if failCount < cb.maxFailures {
		return true
	}

	// In open state - check if we can transition to half-open
	lastAttempt, exists := cb.lastAttempt[domain]
	if !exists || time.Since(lastAttempt) >= cb.retryDelay {
		// Allow probe (half-open state)
		return true
	}

	// Still in open state, retry delay not elapsed
	return false
}

func (cb *circuitBreaker) recordFailure(domain string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures[domain]++
	cb.successes[domain] = 0 // Reset success count
	cb.lastAttempt[domain] = time.Now()
}

func (cb *circuitBreaker) recordSuccess(domain string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	failCount := cb.failures[domain]

	// If in open or half-open state
	if failCount >= cb.maxFailures {
		cb.successes[domain]++

		// If we've reached the success threshold, reset to closed state
		if cb.successes[domain] >= cb.successThreshold {
			cb.failures[domain] = 0
			cb.successes[domain] = 0
			delete(cb.lastAttempt, domain)
		}
	}
}