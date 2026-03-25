package observability

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
)

func TestTracerProvider(t *testing.T) {
	tp, err := NewTracerProvider("test-service", "1.0.0")
	require.NoError(t, err)
	require.NotNil(t, tp)

	// Test tracer creation
	tracer := tp.Tracer("test")
	require.NotNil(t, tracer)

	// Test span creation
	ctx, span := StartSpan(context.Background(), tracer, "test-span")
	require.NotNil(t, ctx)
	require.NotNil(t, span)

	span.End()

	// Test span from context
	retrievedSpan := SpanFromContext(ctx)
	assert.NotNil(t, retrievedSpan)

	// Test shutdown
	err = tp.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestSpanProcessor(t *testing.T) {
	sp := NewSpanProcessor()
	require.NotNil(t, sp)

	tp, err := NewTracerProvider("test-service", "1.0.0")
	require.NoError(t, err)
	defer func() { _ = tp.Shutdown(context.Background()) }()

	tracer := tp.Tracer("test")
	ctx, span := StartSpan(context.Background(), tracer, "test-span")
	defer span.End()

	// Test adding attributes
	sp.AddAttributes(ctx, attribute.String("test.key", "test.value"))

	// Test setting status
	sp.SetStatus(ctx, 0, "ok") // OK status

	// Test recording error
	testErr := assert.AnError
	sp.RecordError(ctx, testErr)

	// Test adding event
	sp.AddEvent(ctx, "test-event")
}
