package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JoX23/go-without-magic/internal/observability"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTracingMiddleware(t *testing.T) {
	// Setup observability
	tp, err := observability.NewTracerProvider("test-service", "1.0.0")
	require.NoError(t, err)
	defer func() { _ = tp.Shutdown(context.Background()) }()

	metrics := observability.NewMetricsWithRegistry(prometheus.NewRegistry())
	spanProcessor := observability.NewSpanProcessor()
	tracer := tp.Tracer("test")

	// Create middleware
	tracingMW := Tracing(tracer, spanProcessor, metrics)

	// Create test handler
	handler := tracingMW(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test response", w.Body.String())
}

func TestBusinessMetricsMiddleware(t *testing.T) {
	metrics := observability.NewMetricsWithRegistry(prometheus.NewRegistry())
	spanProcessor := observability.NewSpanProcessor()

	// Create middleware
	businessMW := BusinessMetrics(metrics, spanProcessor)

	// Create test handler for user creation
	handler := businessMW(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("user created"))
	}))

	// Test user creation endpoint
	req := httptest.NewRequest("POST", "/users", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "user created", w.Body.String())
}

func TestTracingResponseWriter(t *testing.T) {
	tp, err := observability.NewTracerProvider("test-service", "1.0.0")
	require.NoError(t, err)
	defer func() { _ = tp.Shutdown(context.Background()) }()

	metrics := observability.NewMetricsWithRegistry(prometheus.NewRegistry())
	spanProcessor := observability.NewSpanProcessor()
	tracer := tp.Tracer("test")

	ctx, span := observability.StartSpan(context.Background(), tracer, "test")
	defer span.End()

	// Use metrics and ctx to avoid unused variable errors
	_ = metrics
	_ = ctx

	// Create tracing response writer
	w := httptest.NewRecorder()
	tracingWriter := &tracingResponseWriter{
		ResponseWriter: w,
		span:           span,
		spanProcessor:  spanProcessor,
	}

	// Test writing header
	tracingWriter.WriteHeader(http.StatusOK)
	assert.Equal(t, http.StatusOK, tracingWriter.statusCode)

	// Test writing data
	data := []byte("test")
	n, err := tracingWriter.Write(data)
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
}
