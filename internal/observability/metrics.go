package observability

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metrics
type Metrics struct {
	// HTTP metrics
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight *prometheus.GaugeVec

	// Business metrics
	userOperationsTotal *prometheus.CounterVec

	// System metrics
	uptime *prometheus.Counter
}

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics() *Metrics {
	m := &Metrics{
		// HTTP request counter
		httpRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),

		// HTTP request duration histogram
		httpRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets, // 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10
			},
			[]string{"method", "endpoint"},
		),

		// HTTP requests in flight gauge
		httpRequestsInFlight: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Number of HTTP requests currently being processed",
			},
			[]string{"method", "endpoint"},
		),

		// Business metrics - user operations
		userOperationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "user_operations_total",
				Help: "Total number of user operations",
			},
			[]string{"operation", "result"}, // create, get, list + success, error
		),

		// Uptime counter
		uptime: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "service_uptime_seconds_total",
				Help: "Total service uptime in seconds",
			},
		),
	}

	return m
}

// RecordHTTPRequest records an HTTP request metric
func (m *Metrics) RecordHTTPRequest(method, endpoint string, statusCode int, duration time.Duration) {
	statusCodeStr := strconv.Itoa(statusCode)

	m.httpRequestsTotal.WithLabelValues(method, endpoint, statusCodeStr).Inc()
	m.httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// IncrementRequestsInFlight increments the gauge for requests in flight
func (m *Metrics) IncrementRequestsInFlight(method, endpoint string) {
	m.httpRequestsInFlight.WithLabelValues(method, endpoint).Inc()
}

// DecrementRequestsInFlight decrements the gauge for requests in flight
func (m *Metrics) DecrementRequestsInFlight(method, endpoint string) {
	m.httpRequestsInFlight.WithLabelValues(method, endpoint).Dec()
}

// RecordUserOperation records a user operation metric
func (m *Metrics) RecordUserOperation(operation, result string) {
	m.userOperationsTotal.WithLabelValues(operation, result).Inc()
}

// RecordUptime records service uptime
func (m *Metrics) RecordUptime(seconds float64) {
	m.uptime.Add(seconds)
}

// Handler returns the Prometheus metrics HTTP handler
func (m *Metrics) Handler() http.Handler {
	return promhttp.Handler()
}