package observability

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// SpanProcessor provides utilities for working with OpenTelemetry spans
type SpanProcessor struct{}

// NewSpanProcessor creates a new span processor
func NewSpanProcessor() *SpanProcessor {
	return &SpanProcessor{}
}

// AddAttributes adds attributes to the current span
func (sp *SpanProcessor) AddAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := SpanFromContext(ctx)
	if span != nil {
		span.SetAttributes(attrs...)
	}
}

// SetStatus sets the status of the current span
func (sp *SpanProcessor) SetStatus(ctx context.Context, code codes.Code, description string) {
	span := SpanFromContext(ctx)
	if span != nil {
		span.SetStatus(code, description)
	}
}

// RecordError records an error on the current span
func (sp *SpanProcessor) RecordError(ctx context.Context, err error, opts ...trace.EventOption) {
	span := SpanFromContext(ctx)
	if span != nil {
		span.RecordError(err, opts...)
		span.SetStatus(codes.Error, err.Error())
	}
}

// AddEvent adds an event to the current span
func (sp *SpanProcessor) AddEvent(ctx context.Context, name string, opts ...trace.EventOption) {
	span := SpanFromContext(ctx)
	if span != nil {
		span.AddEvent(name, opts...)
	}
}

// HTTPAttributes returns common HTTP attributes for spans
func HTTPAttributes(method, url, userAgent string, statusCode int) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("http.method", method),
		attribute.String("http.url", url),
		attribute.String("http.user_agent", userAgent),
		attribute.Int("http.status_code", statusCode),
	}
}

// DatabaseAttributes returns common database attributes for spans
func DatabaseAttributes(operation, table string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("db.operation", operation),
		attribute.String("db.table", table),
	}
}

// BusinessAttributes returns business logic attributes for spans
func BusinessAttributes(operation, entityType, entityID string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("business.operation", operation),
		attribute.String("business.entity_type", entityType),
		attribute.String("business.entity_id", entityID),
	}
}
