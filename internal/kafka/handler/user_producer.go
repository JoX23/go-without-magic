package kafkahandler

import (
	"context"
	"encoding/json"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"

	"github.com/JoX23/go-without-magic/internal/domain"
	"github.com/JoX23/go-without-magic/internal/kafka"
)

// UserEventProducer emite domain events de usuarios tras llamadas exitosas al servicio.
type UserEventProducer struct {
	producer *kafka.Producer
	logger   *zap.Logger
}

func NewUserEventProducer(p *kafka.Producer, logger *zap.Logger) *UserEventProducer {
	return &UserEventProducer{producer: p, logger: logger}
}

type userCreatedEvent struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// EmitUserCreated publica un evento "user.events.created" con el ID como key
// para garantizar que todos los eventos del mismo usuario van a la misma partición.
func (p *UserEventProducer) EmitUserCreated(ctx context.Context, user *domain.User) error {
	payload, err := json.Marshal(userCreatedEvent{
		ID:    user.ID.String(),
		Email: user.Email,
		Name:  user.Name,
	})
	if err != nil {
		return err
	}

	if err := p.producer.Produce(ctx, &kgo.Record{
		Topic: "user.events.created",
		Key:   []byte(user.ID.String()),
		Value: payload,
	}); err != nil {
		p.logger.Error("failed to emit user created event",
			zap.String("id", user.ID.String()),
			zap.Error(err))
		return err
	}

	p.logger.Info("user created event emitted", zap.String("id", user.ID.String()))
	return nil
}
