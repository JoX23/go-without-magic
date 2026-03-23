package resilience

import (
	"errors"
	"net/http"
	"sync"
	"time"
)

var ErrCircuitBreakerOpen = errors.New("circuit breaker is open")

type CircuitBreaker struct {
	mu               sync.Mutex
	state            int
	consecutiveFails int
	maxFails         int
	timeout          time.Duration
	lastFailTime     time.Time
}

func NewCircuitBreaker(maxFails int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:    0,
		maxFails: maxFails,
		timeout:  timeout,
	}
}

func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()
	state := cb.state
	cb.mu.Unlock()

	switch state {
	case 0:
		err := fn()
		if err != nil {
			cb.recordFail()
		} else {
			cb.recordSuccess()
		}
		return err
	case 1:
		cb.mu.Lock()
		elapsed := time.Since(cb.lastFailTime)
		cb.mu.Unlock()

		if elapsed > cb.timeout {
			cb.mu.Lock()
			cb.state = 2
			cb.mu.Unlock()
			return nil
		}
		return ErrCircuitBreakerOpen
	case 2:
		err := fn()
		if err == nil {
			cb.recordSuccess()
		} else {
			cb.recordFail()
		}
		return err
	}

	return fn()
}

func (cb *CircuitBreaker) recordFail() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.consecutiveFails++
	if cb.consecutiveFails >= cb.maxFails {
		cb.state = 1
		cb.lastFailTime = time.Now()
	}
}

func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.consecutiveFails = 0
	if cb.state == 2 {
		cb.state = 0
	}
}

func (cb *CircuitBreaker) HTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := cb.Call(func() error {
				next.ServeHTTP(w, r)
				return nil
			})

			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte(`{"error":"service unavailable"}`))
			}
		})
	}
}

type RateLimiter struct {
	mu         sync.Mutex
	tokens     float64
	maxTokens  float64
	refillRate float64
	lastRefill time.Time
}

func NewRateLimiter(requestsPerSecond float64) *RateLimiter {
	return &RateLimiter{
		maxTokens:  requestsPerSecond,
		tokens:     requestsPerSecond,
		refillRate: requestsPerSecond,
		lastRefill: time.Now(),
	}
}

func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.tokens = minFloat(rl.maxTokens, rl.tokens+elapsed*rl.refillRate)
	rl.lastRefill = now

	if rl.tokens >= 1.0 {
		rl.tokens -= 1.0
		return true
	}

	return false
}

func (rl *RateLimiter) HTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !rl.Allow() {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "1")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"rate limit exceeded"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
