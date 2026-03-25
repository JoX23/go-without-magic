package kafka

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

// KafkaMetrics agrupa las métricas Prometheus específicas de Kafka.
// Se registran solo cuando el transporte está activo (brokers configurados),
// por eso vive aquí y no en internal/observability/metrics.go.
type KafkaMetrics struct {
	messagesProcessed *prometheus.CounterVec
	retries           *prometheus.CounterVec
	producedMessages  *prometheus.CounterVec
	produceErrors     *prometheus.CounterVec
	consumerLag       *prometheus.GaugeVec
	processingTime    *prometheus.HistogramVec
}

// NewKafkaMetrics registra métricas contra el Registerer global de Prometheus.
func NewKafkaMetrics() *KafkaMetrics {
	return newKafkaMetrics(prometheus.DefaultRegisterer)
}

func newKafkaMetrics(reg prometheus.Registerer) *KafkaMetrics {
	m := &KafkaMetrics{
		messagesProcessed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "kafka_messages_processed_total",
			Help: "Total de mensajes Kafka procesados.",
		}, []string{"topic", "result"}),

		retries: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "kafka_message_retries_total",
			Help: "Total de reintentos de procesamiento de mensajes Kafka.",
		}, []string{"topic"}),

		producedMessages: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "kafka_messages_produced_total",
			Help: "Total de mensajes Kafka producidos.",
		}, []string{"topic"}),

		produceErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "kafka_produce_errors_total",
			Help: "Total de errores al producir mensajes Kafka.",
		}, []string{"topic"}),

		consumerLag: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "kafka_consumer_lag",
			Help: "Lag del consumer group por topic/partición.",
		}, []string{"topic", "partition"}),

		processingTime: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "kafka_message_processing_duration_seconds",
			Help:    "Duración del procesamiento de mensajes Kafka.",
			Buckets: prometheus.DefBuckets,
		}, []string{"topic"}),
	}

	reg.MustRegister(
		m.messagesProcessed,
		m.retries,
		m.producedMessages,
		m.produceErrors,
		m.consumerLag,
		m.processingTime,
	)
	return m
}

func (m *KafkaMetrics) RecordMessage(topic, result string) {
	m.messagesProcessed.WithLabelValues(topic, result).Inc()
}

func (m *KafkaMetrics) RecordRetry(topic string) {
	m.retries.WithLabelValues(topic).Inc()
}

func (m *KafkaMetrics) RecordProduced(topic string) {
	m.producedMessages.WithLabelValues(topic).Inc()
}

func (m *KafkaMetrics) RecordProduceError(topic string) {
	m.produceErrors.WithLabelValues(topic).Inc()
}

func (m *KafkaMetrics) RecordFetchError(topic string) {
	m.messagesProcessed.WithLabelValues(topic, "fetch_error").Inc()
}

func (m *KafkaMetrics) UpdateConsumerLag(topic string, partition int32, lag float64) {
	m.consumerLag.WithLabelValues(topic, fmt.Sprintf("%d", partition)).Set(lag)
}

func (m *KafkaMetrics) ObserveProcessingTime(topic string, seconds float64) {
	m.processingTime.WithLabelValues(topic).Observe(seconds)
}
