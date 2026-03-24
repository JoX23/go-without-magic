package domain

import "context"

// ProductRepository define el contrato del puerto de salida.
//
// El dominio define la interfaz; la implementación vive en
// internal/repository/. Esto es el patrón Port & Adapter.
//
// REGLA: esta interfaz NO importa nada fuera del paquete domain.
type ProductRepository interface {
	// CreateIfNotExists crea la entidad si no existe un duplicado.
	// Retorna ErrProductDuplicated si ya existe.
	//
	// GARANTÍA: Thread-safe. Operación atómica.
	CreateIfNotExists(ctx context.Context, e *Product) error

	// Save crea o actualiza la entidad (incondicionalmente).
	Save(ctx context.Context, e *Product) error

	// FindByID busca por ID.
	// Retorna ErrProductNotFound si no existe.
	FindByID(ctx context.Context, id string) (*Product, error)

	// FindBySku busca por Sku.
	// Retorna ErrProductNotFound si no existe.
	FindBySku(ctx context.Context, sku string) (*Product, error)

	// List retorna todos los registros.
	List(ctx context.Context) ([]*Product, error)
}
