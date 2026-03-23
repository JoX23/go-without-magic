package resilience_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JoX23/go-without-magic/internal/resilience"
)

func TestCircuitBreakerClosed(t *testing.T) {
	cb := resilience.NewCircuitBreaker(3, 100*time.Millisecond)

	for i := 0; i < 5; i++ {
		err := cb.Call(func() error { return nil })
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}
}

func TestCircuitBreakerOpen(t *testing.T) {
	cb := resilience.NewCircuitBreaker(2, 100*time.Millisecond)

	fn := func() error { return errors.New("error") }

	// Trigger failures
	cb.Call(fn)
	cb.Call(fn)

	// Should be open now
	err := cb.Call(fn)
	if !errors.Is(err, resilience.ErrCircuitBreakerOpen) {
		t.Fatalf("expected circuit breaker open, got %v", err)
	}
}

func TestRateLimiterAllow(t *testing.T) {
	rl := resilience.NewRateLimiter(10)

	allowed := 0
	for i := 0; i < 20; i++ {
		if rl.Allow() {
			allowed++
		}
	}

	if allowed != 10 {
		t.Fatalf("expected 10 allowed, got %d", allowed)
	}
}

func TestRateLimiterRefill(t *testing.T) {
	rl := resilience.NewRateLimiter(10)

	// Consume all
	for i := 0; i < 10; i++ {
		if !rl.Allow() {
			t.Fatalf("token available at %d", i)
		}
	}

	// Should fail
	if rl.Allow() {
		t.Fatal("expected no tokens")
	}

	// Wait for refill
	time.Sleep(150 * time.Millisecond)

	// Should have token now
	if !rl.Allow() {
		t.Fatal("expected refill")
	}
}

func TestRateLimiterMiddleware(t *testing.T) {
	rl := resilience.NewRateLimiter(2)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := rl.HTTPMiddleware()
	h := mw(handler)

	// First two should pass
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		h.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 at request %d, got %d", i, w.Code)
		}
	}

	// Third should fail
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	h.ServeHTTP(w, r)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}
}
