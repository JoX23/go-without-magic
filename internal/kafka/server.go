package kafka

import "context"

// KafkaConsumerAdapter implementa shutdown.Server para el Consumer.
// Espejo exacto de GRPCServerAdapter en internal/grpc/server.go.
type KafkaConsumerAdapter struct {
	Consumer *Consumer
}

func (a *KafkaConsumerAdapter) Shutdown(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		// Close() es idempotente y hace que PollFetches retorne inmediatamente.
		a.Consumer.client.Close()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
