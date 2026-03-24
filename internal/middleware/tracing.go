package middleware

import (
	"net/http"
	"time"

	"github.com/JoX23/go-without-magic/internal/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// Tracing agrega OpenTelemetry tracing a requests HTTP.
func Tracing(tracer trace.Tracer, spanProcessor *observability.SpanProcessor, metrics *observability.Metrics) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Start span for this request
			ctx, span := tracer.Start(r.Context(), "http.request",
				trace.WithAttributes(
					semconv.HTTPMethodKey.String(r.Method),
					semconv.HTTPURLKey.String(r.URL.String()),
					semconv.HTTPUserAgentKey.String(r.Header.Get("User-Agent")),
				),
			)
			defer span.End()

			// Wrap response writer to capture status code
			wrapped := &tracingResponseWriter{
				ResponseWriter: w,
				span:           span,
				spanProcessor:  spanProcessor,
				startTime:      time.Now(),
			}

			// Add metrics tracking
			metrics.IncrementRequestsInFlight(r.Method, r.URL.Path)
			defer func() {
				duration := time.Since(wrapped.startTime)
				metrics.DecrementRequestsInFlight(r.Method, r.URL.Path)
				metrics.RecordHTTPRequest(r.Method, r.URL.Path, wrapped.statusCode, duration)
			}()

			// Call next handler with tracing context
			next.ServeHTTP(wrapped, r.WithContext(ctx))
		})
	}
}

// tracingResponseWriter captura el status code y actualiza el span.
type tracingResponseWriter struct {
	http.ResponseWriter
	span          trace.Span
	spanProcessor *observability.SpanProcessor
	statusCode    int
	startTime     time.Time
	written       bool
}

func (w *tracingResponseWriter) WriteHeader(code int) {
	if !w.written {
		w.statusCode = code
		w.written = true

		// Update span with status code
		w.span.SetAttributes(attribute.Int("http.status_code", code))

		// Set span status based on HTTP status
		if code >= 400 {
			w.span.SetStatus(codes.Error, http.StatusText(code))
		} else {
			w.span.SetStatus(codes.Ok, "")
		}

		w.ResponseWriter.WriteHeader(code)
	}
}

func (w *tracingResponseWriter) Write(data []byte) (int, error) {
	if !w.written {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(data)
}
