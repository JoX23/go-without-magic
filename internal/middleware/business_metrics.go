package middleware

import (
	"net/http"
	"time"

	"github.com/JoX23/go-without-magic/internal/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// BusinessMetrics registra métricas de negocio para operaciones específicas.
func BusinessMetrics(metrics *observability.Metrics, spanProcessor *observability.SpanProcessor) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture business operation details
			wrapped := &businessMetricsWriter{
				ResponseWriter: w,
				metrics:        metrics,
				spanProcessor:  spanProcessor,
				startTime:      start,
				path:           r.URL.Path,
				method:         r.Method,
			}

			next.ServeHTTP(wrapped, r)

			// Record business operation if it was identified
			if wrapped.operation != "" {
				duration := time.Since(start)
				result := "success"
				if wrapped.statusCode >= 400 {
					result = "error"
				}

				metrics.RecordUserOperation(wrapped.operation, result)

				// Add business attributes to span
				span := trace.SpanFromContext(r.Context())
				if span != nil {
					span.SetAttributes(
						attribute.String("business.operation", wrapped.operation),
						attribute.String("business.result", result),
						attribute.Int64("business.duration_ms", duration.Milliseconds()),
					)
				}
			}
		})
	}
}

// businessMetricsWriter captura detalles de operaciones de negocio.
type businessMetricsWriter struct {
	http.ResponseWriter
	metrics       *observability.Metrics
	spanProcessor *observability.SpanProcessor
	startTime     time.Time
	path          string
	method        string
	statusCode    int
	operation     string // e.g., "user.create", "user.get", "user.list"
	written       bool
}

func (w *businessMetricsWriter) WriteHeader(code int) {
	if !w.written {
		w.statusCode = code
		w.written = true

		// Identify business operation based on path
		w.operation = w.identifyOperation(w.path, w.method)

		w.ResponseWriter.WriteHeader(code)
	}
}

func (w *businessMetricsWriter) Write(data []byte) (int, error) {
	if !w.written {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(data)
}

// identifyOperation determina la operación de negocio basada en path y método.
func (w *businessMetricsWriter) identifyOperation(path, method string) string {
	switch {
	case path == "/users" && method == "POST":
		return "user.create"
	case path == "/users" && method == "GET":
		return "user.list"
	case path == "/users/" && method == "GET": // Note: this is a simple check
		return "user.get"
	default:
		return ""
	}
}
