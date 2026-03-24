package memory

import (
	"context"
	"sync"

	"github.com/JoX23/go-without-magic/internal/domain"
)

// ProductRepository es una implementación en memoria del dominio.
//
// Usos:
//   - Tests unitarios y de integración (sin base de datos real)
//   - Desarrollo local sin infraestructura
//
// Es seguro para uso concurrente (sync.RWMutex).
type ProductRepository struct {
	mu    sync.RWMutex
	byID  map[string]*domain.Product
	bySku map[string]*domain.Product
}

func NewProductRepository() *ProductRepository {
	return &ProductRepository{
		byID:  make(map[string]*domain.Product),
		bySku: make(map[string]*domain.Product),
	}
}

// CreateIfNotExists verifica unicidad y crea de forma ATÓMICA.
// Retorna ErrProductDuplicated si ya existe un duplicado.
//
// GARANTÍA DE CONCURRENCIA: Thread-safe, sin ventana entre check y write.
func (r *ProductRepository) CreateIfNotExists(_ context.Context, e *domain.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.bySku[e.Sku]; exists {
		return domain.ErrProductDuplicated
	}

	r.byID[e.ID.String()] = e
	r.bySku[e.Sku] = e

	return nil
}

func (r *ProductRepository) Save(_ context.Context, e *domain.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.byID[e.ID.String()] = e
	r.bySku[e.Sku] = e

	return nil
}

func (r *ProductRepository) FindByID(_ context.Context, id string) (*domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	e, ok := r.byID[id]
	if !ok {
		return nil, domain.ErrProductNotFound
	}
	return e, nil
}

func (r *ProductRepository) FindBySku(_ context.Context, sku string) (*domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	e, ok := r.bySku[sku]
	if !ok {
		return nil, domain.ErrProductNotFound
	}
	return e, nil
}

func (r *ProductRepository) List(_ context.Context) ([]*domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]*domain.Product, 0, len(r.byID))
	for _, e := range r.byID {
		items = append(items, e)
	}
	return items, nil
}
