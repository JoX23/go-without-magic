package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"

	"github.com/JoX23/go-without-magic/internal/config"
)

// Consumer gestiona el loop de consumo Kafka con retry y Dead Letter Topic.
type Consumer struct {
	client   *kgo.Client
	admin    *kadm.Client
	handlers TopicHandlerMap
	producer *Producer
	metrics  *KafkaMetrics
	cfg      config.KafkaConfig
	logger   *zap.Logger
}

// NewConsumer crea el cliente franz-go con consumer group y commit manual.
func NewConsumer(
	cfg config.KafkaConfig,
	handlers TopicHandlerMap,
	producer *Producer,
	metrics *KafkaMetrics,
	logger *zap.Logger,
) (*Consumer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ConsumerGroup(cfg.ConsumerGroup),
		kgo.ConsumeTopics(cfg.Topics...),
		kgo.SessionTimeout(cfg.SessionTimeout),
		kgo.RebalanceTimeout(cfg.RebalanceTimeout),
		kgo.DisableAutoCommit(),
	)
	if err != nil {
		return nil, fmt.Errorf("creating kafka consumer client: %w", err)
	}

	return &Consumer{
		client:   client,
		admin:    kadm.NewClient(client),
		handlers: handlers,
		producer: producer,
		metrics:  metrics,
		cfg:      cfg,
		logger:   logger,
	}, nil
}

// Start lanza el loop de consumo y el reporter de lag en goroutines separadas.
// El loop corre hasta que ctx sea cancelado (vía Shutdown → client.Close()).
func (c *Consumer) Start(ctx context.Context) {
	go c.pollLoop(ctx)
	go c.lagReporter(ctx)
}

func (c *Consumer) pollLoop(ctx context.Context) {
	for {
		fetches := c.client.PollFetches(ctx)
		if ctx.Err() != nil {
			return
		}

		fetches.EachError(func(topic string, partition int32, err error) {
			c.logger.Error("kafka fetch error",
				zap.String("topic", topic),
				zap.Int32("partition", partition),
				zap.Error(err))
			c.metrics.RecordFetchError(topic)
		})

		fetches.EachRecord(func(rec *kgo.Record) {
			start := time.Now()
			c.processWithRetry(ctx, rec)
			c.metrics.ObserveProcessingTime(rec.Topic, time.Since(start).Seconds())
		})

		// Commit solo después de procesar todo el batch (at-least-once).
		if err := c.client.CommitUncommittedOffsets(ctx); err != nil && ctx.Err() == nil {
			c.logger.Error("kafka commit failed", zap.Error(err))
		}
	}
}

func (c *Consumer) processWithRetry(ctx context.Context, rec *kgo.Record) {
	handler, ok := c.handlers[rec.Topic]
	if !ok {
		c.logger.Warn("no handler for topic", zap.String("topic", rec.Topic))
		c.sendToDLT(ctx, rec, fmt.Errorf("no handler registered for topic %q", rec.Topic))
		return
	}

	var lastErr error
	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		err := handler.Handle(ctx, rec)
		if err == nil {
			c.metrics.RecordMessage(rec.Topic, "success")
			return
		}
		lastErr = err

		switch DispositionFor(err) {
		case DispositionSkip:
			c.logger.Info("kafka message skipped (idempotent)",
				zap.String("topic", rec.Topic),
				zap.Error(err))
			c.metrics.RecordMessage(rec.Topic, "skipped")
			return

		case DispositionDLT:
			c.sendToDLT(ctx, rec, err)
			return

		case DispositionRetry:
			if attempt < c.cfg.MaxRetries {
				c.logger.Warn("kafka message retry",
					zap.String("topic", rec.Topic),
					zap.Int("attempt", attempt+1),
					zap.Int("max_retries", c.cfg.MaxRetries),
					zap.Error(err))
				c.metrics.RecordRetry(rec.Topic)
				continue
			}
		}
	}

	// Reintentos agotados → DLT
	c.sendToDLT(ctx, rec, lastErr)
}

func (c *Consumer) sendToDLT(ctx context.Context, rec *kgo.Record, cause error) {
	dltTopic := rec.Topic + c.cfg.DLTSuffix
	c.logger.Error("kafka sending to DLT",
		zap.String("original_topic", rec.Topic),
		zap.String("dlt_topic", dltTopic),
		zap.Error(cause))
	c.metrics.RecordMessage(rec.Topic, "dlt")

	headers := make([]kgo.RecordHeader, len(rec.Headers)+1)
	copy(headers, rec.Headers)
	headers[len(rec.Headers)] = kgo.RecordHeader{
		Key:   "x-dlt-error",
		Value: []byte(cause.Error()),
	}

	if err := c.producer.Produce(ctx, &kgo.Record{
		Topic:   dltTopic,
		Key:     rec.Key,
		Value:   rec.Value,
		Headers: headers,
	}); err != nil {
		c.logger.Error("kafka DLT produce failed",
			zap.String("dlt_topic", dltTopic),
			zap.Error(err))
	}
}

// lagReporter reporta el consumer lag a Prometheus cada 30 segundos.
func (c *Consumer) lagReporter(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			lags, err := c.admin.Lag(ctx, c.cfg.ConsumerGroup)
			if err != nil {
				c.logger.Warn("kafka lag report failed", zap.Error(err))
				continue
			}
			lags.Each(func(l kadm.DescribedGroupLag) {
				for _, m := range l.Lag.Sorted() {
					if m.Err == nil {
						c.metrics.UpdateConsumerLag(m.Topic, m.Partition, float64(m.Lag))
					}
				}
			})
		}
	}
}
