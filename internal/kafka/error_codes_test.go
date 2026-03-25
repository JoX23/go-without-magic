package kafka_test

import (
	"errors"
	"testing"

	"github.com/JoX23/go-without-magic/internal/domain"
	"github.com/JoX23/go-without-magic/internal/kafka"
	"github.com/JoX23/go-without-magic/internal/resilience"
)

func TestDispositionFor(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want kafka.Disposition
	}{
		{
			name: "invalid email → DLT",
			err:  domain.ErrInvalidEmail,
			want: kafka.DispositionDLT,
		},
		{
			name: "invalid name → DLT",
			err:  domain.ErrInvalidName,
			want: kafka.DispositionDLT,
		},
		{
			name: "wrapped invalid email → DLT",
			err:  errors.Join(domain.ErrInvalidEmail, errors.New("extra context")),
			want: kafka.DispositionDLT,
		},
		{
			name: "user duplicated → Skip",
			err:  domain.ErrUserDuplicated,
			want: kafka.DispositionSkip,
		},
		{
			name: "circuit breaker open → Retry",
			err:  resilience.ErrCircuitBreakerOpen,
			want: kafka.DispositionRetry,
		},
		{
			name: "generic error → Retry",
			err:  errors.New("connection timeout"),
			want: kafka.DispositionRetry,
		},
		{
			name: "user not found → Retry",
			err:  domain.ErrUserNotFound,
			want: kafka.DispositionRetry,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := kafka.DispositionFor(tt.err)
			if got != tt.want {
				t.Errorf("DispositionFor(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
