package kafka_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"

	"github.com/JoX23/go-without-magic/internal/kafka"
	"github.com/JoX23/go-without-magic/internal/resilience"
)

func TestWithPanic_RecoversPanic(t *testing.T) {
	logger := zap.NewNop()
	panicHandler := kafka.MessageHandlerFunc(func(_ context.Context, _ *kgo.Record) error {
		panic("unexpected error")
	})

	wrapped := kafka.WithPanic(logger, panicHandler)
	err := wrapped.Handle(context.Background(), &kgo.Record{Topic: "test.topic"})

	if err == nil {
		t.Fatal("expected error from panic recovery, got nil")
	}
}

func TestWithPanic_PassesThroughNilError(t *testing.T) {
	logger := zap.NewNop()
	okHandler := kafka.MessageHandlerFunc(func(_ context.Context, _ *kgo.Record) error {
		return nil
	})

	wrapped := kafka.WithPanic(logger, okHandler)
	if err := wrapped.Handle(context.Background(), &kgo.Record{Topic: "test.topic"}); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestWithPanic_PassesThroughError(t *testing.T) {
	logger := zap.NewNop()
	sentinel := errors.New("domain error")
	errHandler := kafka.MessageHandlerFunc(func(_ context.Context, _ *kgo.Record) error {
		return sentinel
	})

	wrapped := kafka.WithPanic(logger, errHandler)
	err := wrapped.Handle(context.Background(), &kgo.Record{Topic: "test.topic"})

	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestWithCircuitBreaker_OpenReturnsError(t *testing.T) {
	// timeout largo para que no transite a half-open durante el test
	cb := resilience.NewCircuitBreaker(1, time.Hour)

	failHandler := kafka.MessageHandlerFunc(func(_ context.Context, _ *kgo.Record) error {
		return errors.New("downstream error")
	})
	wrapped := kafka.WithCircuitBreaker(cb, failHandler)

	// Primer intento: falla y abre el CB
	_ = wrapped.Handle(context.Background(), &kgo.Record{Topic: "test.topic"})

	// Segundo intento: CB abierto → ErrCircuitBreakerOpen
	err := wrapped.Handle(context.Background(), &kgo.Record{Topic: "test.topic"})
	if !errors.Is(err, resilience.ErrCircuitBreakerOpen) {
		t.Errorf("expected ErrCircuitBreakerOpen, got %v", err)
	}
}

func TestMessageHandlerFunc_ImplementsInterface(t *testing.T) {
	var _ kafka.MessageHandler = kafka.MessageHandlerFunc(func(_ context.Context, _ *kgo.Record) error {
		return nil
	})
}
