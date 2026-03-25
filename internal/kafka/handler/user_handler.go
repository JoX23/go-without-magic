package kafkahandler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"

	"github.com/JoX23/go-without-magic/internal/kafka"
	"github.com/JoX23/go-without-magic/internal/service"
)

// UserKafkaHandler enruta mensajes de los topics user.* hacia UserService.
type UserKafkaHandler struct {
	svc    *service.UserService
	logger *zap.Logger
}

func NewUserKafkaHandler(svc *service.UserService, logger *zap.Logger) *UserKafkaHandler {
	return &UserKafkaHandler{svc: svc, logger: logger}
}

// Register devuelve el TopicHandlerMap para este handler.
// Se llama en main.go y se fusiona con el mapa global antes de crear el Consumer.
func (h *UserKafkaHandler) Register() kafka.TopicHandlerMap {
	return kafka.TopicHandlerMap{
		"user.commands.create": kafka.MessageHandlerFunc(h.handleCreate),
	}
}

type createUserCommand struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (h *UserKafkaHandler) handleCreate(ctx context.Context, msg *kgo.Record) error {
	var cmd createUserCommand
	if err := json.Unmarshal(msg.Value, &cmd); err != nil {
		// Payload mal formado → DLT inmediato (sin reintentos).
		return fmt.Errorf("%w: %v", kafka.ErrInvalidMessage, err)
	}

	user, err := h.svc.CreateUser(ctx, cmd.Email, cmd.Name)
	if err != nil {
		return err // los errores de dominio fluyen a DispositionFor en el consumer
	}

	h.logger.Info("user created via kafka",
		zap.String("id", user.ID.String()),
		zap.String("topic", msg.Topic))
	return nil
}
