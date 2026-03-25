//go:build integration

package kafka_test

// Tests de integración que requieren un broker Kafka real.
// Se ejecutan con: go test -tags=integration -race ./internal/kafka/...
//
// Prerequisito: Docker disponible (usa testcontainers-go).
// O bien, tener Kafka corriendo en localhost:9092 y pasar la variable:
//   KAFKA_BROKERS=localhost:9092 go test -tags=integration ./internal/kafka/...
//
// Para levantar Kafka local:
//   make kafka-up
//
// Ejemplo de test de integración (descomenta cuando agregues testcontainers):
//
// import (
// 	"context"
// 	"encoding/json"
// 	"os"
// 	"testing"
// 	"time"
//
// 	"github.com/twmb/franz-go/pkg/kgo"
// 	"go.uber.org/zap"
//
// 	"github.com/JoX23/go-without-magic/internal/config"
// 	"github.com/JoX23/go-without-magic/internal/kafka"
// )
//
// func TestConsumer_ProcessesMessage_Integration(t *testing.T) {
// 	brokers := os.Getenv("KAFKA_BROKERS")
// 	if brokers == "" {
// 		brokers = "localhost:9092"
// 	}
//
// 	cfg := config.KafkaConfig{
// 		Brokers:          []string{brokers},
// 		ConsumerGroup:    "test-group-" + t.Name(),
// 		Topics:           []string{"user.commands.create.test"},
// 		DLTSuffix:        ".dlt",
// 		MaxRetries:       1,
// 		SessionTimeout:   10 * time.Second,
// 		RebalanceTimeout: 30 * time.Second,
// 	}
//
// 	logger := zap.NewNop()
// 	metrics := kafka.NewKafkaMetrics()
// 	producer, err := kafka.NewProducer(cfg, metrics, logger)
// 	if err != nil {
// 		t.Fatalf("creating producer: %v", err)
// 	}
//
// 	processed := make(chan string, 1)
// 	handler := kafka.MessageHandlerFunc(func(ctx context.Context, msg *kgo.Record) error {
// 		processed <- string(msg.Value)
// 		return nil
// 	})
//
// 	handlers := kafka.TopicHandlerMap{"user.commands.create.test": handler}
// 	consumer, err := kafka.NewConsumer(cfg, handlers, producer, metrics, logger)
// 	if err != nil {
// 		t.Fatalf("creating consumer: %v", err)
// 	}
//
// 	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 	defer cancel()
// 	consumer.Start(ctx)
//
// 	payload, _ := json.Marshal(map[string]string{"email": "test@example.com", "name": "Test"})
// 	if err := producer.Produce(ctx, &kgo.Record{
// 		Topic: "user.commands.create.test",
// 		Value: payload,
// 	}); err != nil {
// 		t.Fatalf("producing message: %v", err)
// 	}
//
// 	select {
// 	case msg := <-processed:
// 		t.Logf("processed message: %s", msg)
// 	case <-ctx.Done():
// 		t.Fatal("timeout waiting for message to be processed")
// 	}
// }
