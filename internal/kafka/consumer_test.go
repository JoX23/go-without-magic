package kafka_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"

	"github.com/JoX23/go-without-magic/internal/domain"
	"github.com/JoX23/go-without-magic/internal/kafka"
)

func TestDispositionFor_SkipOnDuplicate(t *testing.T) {
	calls := 0
	handler := kafka.MessageHandlerFunc(func(_ context.Context, _ *kgo.Record) error {
		calls++
		return domain.ErrUserDuplicated
	})

	// El handler retorna duplicado → DispositionSkip
	// No podemos llamar a processWithRetry directamente (privado),
	// pero sí podemos verificar que DispositionFor clasifica correctamente
	// y que el handler solo se llama una vez en cualquier implementación correcta.
	err := handler.Handle(context.Background(), &kgo.Record{Topic: "user.commands.create"})

	if !errors.Is(err, domain.ErrUserDuplicated) {
		t.Errorf("expected ErrUserDuplicated, got %v", err)
	}
	if calls != 1 {
		t.Errorf("handler should be called once, got %d", calls)
	}
	if kafka.DispositionFor(err) != kafka.DispositionSkip {
		t.Errorf("ErrUserDuplicated should map to DispositionSkip")
	}
}

func TestDispositionFor_DLTOnInvalidInput(t *testing.T) {
	err := domain.ErrInvalidEmail
	if kafka.DispositionFor(err) != kafka.DispositionDLT {
		t.Errorf("ErrInvalidEmail should map to DispositionDLT")
	}
}

func TestTopicHandlerMap_RoutesMissing(t *testing.T) {
	handlers := kafka.TopicHandlerMap{
		"topic.a": kafka.MessageHandlerFunc(func(_ context.Context, _ *kgo.Record) error {
			return nil
		}),
	}

	_, ok := handlers["topic.unknown"]
	if ok {
		t.Error("expected no handler for unknown topic")
	}
}

func TestWithPanic_ConcurrentSafe(t *testing.T) {
	logger := zap.NewNop()
	var calls atomic.Int64

	handler := kafka.WithPanic(logger, kafka.MessageHandlerFunc(func(_ context.Context, _ *kgo.Record) error {
		calls.Add(1)
		return nil
	}))

	ctx := context.Background()
	rec := &kgo.Record{Topic: "concurrent.topic"}

	done := make(chan struct{})
	for i := 0; i < 100; i++ {
		go func() {
			_ = handler.Handle(ctx, rec)
			done <- struct{}{}
		}()
	}
	for i := 0; i < 100; i++ {
		<-done
	}

	if calls.Load() != 100 {
		t.Errorf("expected 100 calls, got %d", calls.Load())
	}
}
