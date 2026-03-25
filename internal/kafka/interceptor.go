package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/JoX23/go-without-magic/internal/resilience"
)

// WithPanic recupera panics del handler y los convierte en error.
// Espejo de UnaryServerInterceptor en internal/grpc/interceptor.go.
func WithPanic(logger *zap.Logger, next MessageHandler) MessageHandler {
	return MessageHandlerFunc(func(ctx context.Context, msg *kgo.Record) (err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("kafka handler panic recovered",
					zap.String("topic", msg.Topic),
					zap.Int32("partition", msg.Partition),
					zap.Any("panic", r))
				err = fmt.Errorf("panic recovered processing topic %s: %v", msg.Topic, r)
			}
		}()
		return next.Handle(ctx, msg)
	})
}

// WithTracing extrae el contexto W3C TraceContext de los headers del mensaje
// y crea un span hijo para el procesamiento.
func WithTracing(tracer trace.Tracer, next MessageHandler) MessageHandler {
	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	return MessageHandlerFunc(func(ctx context.Context, msg *kgo.Record) error {
		ctx = prop.Extract(ctx, kafkaHeaderCarrier(msg.Headers))

		ctx, span := tracer.Start(ctx, "kafka.consume",
			trace.WithSpanKind(trace.SpanKindConsumer),
			trace.WithAttributes(
				semconv.MessagingSystem("kafka"),
				semconv.MessagingDestinationName(msg.Topic),
				attribute.Int("messaging.kafka.partition", int(msg.Partition)),
				attribute.Int64("messaging.kafka.offset", msg.Offset),
			),
		)
		defer span.End()

		start := time.Now()
		err := next.Handle(ctx, msg)
		_ = start // disponible para extender métricas de span si se desea

		if err != nil {
			span.RecordError(err)
		}
		return err
	})
}

// WithCircuitBreaker envuelve el handler con el CircuitBreaker existente.
// Reutiliza resilience.CircuitBreaker sin modificarlo.
func WithCircuitBreaker(cb *resilience.CircuitBreaker, next MessageHandler) MessageHandler {
	return MessageHandlerFunc(func(ctx context.Context, msg *kgo.Record) error {
		return cb.Call(func() error {
			return next.Handle(ctx, msg)
		})
	})
}

// kafkaHeaderCarrier adapta []kgo.RecordHeader a otel TextMapCarrier (solo lectura).
type kafkaHeaderCarrier []kgo.RecordHeader

func (c kafkaHeaderCarrier) Get(key string) string {
	for _, h := range c {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

func (c kafkaHeaderCarrier) Set(string, string) {}

func (c kafkaHeaderCarrier) Keys() []string {
	keys := make([]string, len(c))
	for i, h := range c {
		keys[i] = h.Key
	}
	return keys
}
