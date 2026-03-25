package kafka

import (
	"errors"

	"github.com/JoX23/go-without-magic/internal/domain"
	"github.com/JoX23/go-without-magic/internal/resilience"
)

// ErrInvalidMessage indica que el payload del mensaje no se puede parsear.
// Mapea a DispositionDLT: reintentar no arregla un mensaje mal formado.
var ErrInvalidMessage = errors.New("invalid kafka message payload")

// Disposition determina qué hace el consumer con un mensaje que falló.
type Disposition int

const (
	// DispositionRetry reintenta el mensaje hasta cfg.MaxRetries, luego DLT.
	DispositionRetry Disposition = iota
	// DispositionDLT envía al Dead Letter Topic inmediatamente (sin reintentos).
	DispositionDLT
	// DispositionSkip commit el offset y lo trata como éxito (idempotencia).
	DispositionSkip
)

// DispositionFor mapea un error al tratamiento correcto.
// Espejo de internal/grpc/error_codes.go: dominio → protocolo de transporte.
func DispositionFor(err error) Disposition {
	switch {
	// Mensaje mal formado (JSON inválido, campos faltantes) → DLT inmediato.
	case errors.Is(err, ErrInvalidMessage):
		return DispositionDLT

	// Errores de validación de dominio: reintentar no repara input inválido → DLT inmediato.
	case errors.Is(err, domain.ErrInvalidEmail),
		errors.Is(err, domain.ErrInvalidName):
		return DispositionDLT

	// Duplicado: el mensaje ya fue procesado → skip (at-least-once idempotente).
	case errors.Is(err, domain.ErrUserDuplicated):
		return DispositionSkip

	// Circuit breaker abierto: downstream no disponible → reintentar.
	case errors.Is(err, resilience.ErrCircuitBreakerOpen):
		return DispositionRetry

	default:
		return DispositionRetry
	}
}
