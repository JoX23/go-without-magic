package kafka

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"

	"github.com/JoX23/go-without-magic/internal/config"
)

// Producer emite mensajes a Kafka con garantía at-least-once (AllISRAcks).
// Usado por los handlers de negocio para emitir domain events,
// y por el Consumer para enrutar mensajes al Dead Letter Topic.
type Producer struct {
	client  *kgo.Client
	metrics *KafkaMetrics
	logger  *zap.Logger
}

// NewProducer crea un cliente productor con particionamiento por key
// y compresión Snappy. Requiere ack de todas las réplicas ISR.
func NewProducer(cfg config.KafkaConfig, metrics *KafkaMetrics, logger *zap.Logger) (*Producer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.RecordPartitioner(kgo.StickyKeyPartitioner(nil)),
		kgo.ProducerBatchCompression(kgo.SnappyCompression()),
		kgo.RequiredAcks(kgo.AllISRAcks()),
	)
	if err != nil {
		return nil, fmt.Errorf("creating kafka producer client: %w", err)
	}
	return &Producer{client: client, metrics: metrics, logger: logger}, nil
}

// Produce envía un registro de forma síncrona y espera el ack del broker.
func (p *Producer) Produce(ctx context.Context, rec *kgo.Record) error {
	if err := p.client.ProduceSync(ctx, rec).FirstErr(); err != nil {
		p.logger.Error("kafka produce failed",
			zap.String("topic", rec.Topic),
			zap.Error(err))
		p.metrics.RecordProduceError(rec.Topic)
		return fmt.Errorf("producing to %s: %w", rec.Topic, err)
	}
	p.metrics.RecordProduced(rec.Topic)
	return nil
}

// Shutdown flush los registros pendientes y cierra el cliente.
func (p *Producer) Shutdown(ctx context.Context) error {
	p.client.Flush(ctx)
	p.client.Close()
	return nil
}
