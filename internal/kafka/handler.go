package kafka

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"
)

// MessageHandler procesa un único mensaje Kafka.
// Devolver error no-nil activa la pipeline de retry/DLT.
type MessageHandler interface {
	Handle(ctx context.Context, msg *kgo.Record) error
}

// MessageHandlerFunc es el adaptador función, análogo a http.HandlerFunc.
type MessageHandlerFunc func(ctx context.Context, msg *kgo.Record) error

func (f MessageHandlerFunc) Handle(ctx context.Context, msg *kgo.Record) error {
	return f(ctx, msg)
}

// TopicHandlerMap enruta nombres de topic a su handler.
// Se construye en main.go y se pasa al Consumer.
type TopicHandlerMap map[string]MessageHandler
