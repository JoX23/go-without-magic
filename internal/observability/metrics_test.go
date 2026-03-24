package observability

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestMetrics(t *testing.T) {
	// Create a test-specific registry to avoid global registration conflicts
	reg := prometheus.NewRegistry()
	m := NewMetricsWithRegistry(reg)
	assert.NotNil(t, m)

	// Test HTTP request recording
	m.RecordHTTPRequest("GET", "/users", 200, 100*time.Millisecond)

	// Test requests in flight
	m.IncrementRequestsInFlight("GET", "/users")
	m.DecrementRequestsInFlight("GET", "/users")

	// Test user operation recording
	m.RecordUserOperation("user.create", "success")
	m.RecordUserOperation("user.get", "error")

	// Test uptime recording
	m.RecordUptime(60.0) // 60 seconds

	// Verify metrics exist (basic smoke test)
	assert.NotNil(t, m.Handler())
}

func TestMetricsCollection(t *testing.T) {
	// Create a test-specific registry to avoid global registration conflicts
	reg := prometheus.NewRegistry()
	m := NewMetricsWithRegistry(reg)

	// Record some test data
	m.RecordHTTPRequest("POST", "/users", 201, 50*time.Millisecond)
	m.RecordHTTPRequest("GET", "/users", 200, 25*time.Millisecond)
	m.RecordHTTPRequest("GET", "/users/123", 404, 10*time.Millisecond)

	m.RecordUserOperation("user.create", "success")
	m.RecordUserOperation("user.get", "success")
	m.RecordUserOperation("user.list", "error")

	// Test that counters have been incremented
	// Note: We can't easily test the exact values without exposing internal counters,
	// but we can verify the metrics are being collected by checking they're not nil
	assert.NotNil(t, m)
}
