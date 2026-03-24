package observability
package observability

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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























}	sp.AddEvent(ctx, "test-event")	// Test adding event	sp.RecordError(ctx, testErr)	testErr := assert.AnError	// Test recording error	sp.SetStatus(ctx, 0, "ok") // OK status	// Test setting status	sp.AddAttributes(ctx, "test.key", "test.value")	// Test adding attributes	defer span.End()	ctx, span := StartSpan(context.Background(), tracer, "test-span")	tracer := tp.Tracer("test")	defer tp.Shutdown(context.Background())	require.NoError(t, err)	tp, err := NewTracerProvider("test-service", "1.0.0")	require.NotNil(t, sp)	sp := NewSpanProcessor()