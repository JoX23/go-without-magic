package service

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/JoX23/go-without-magic/internal/domain"
)

// ProductService contiene SOLO lógica de negocio.
// No sabe nada de HTTP, gRPC, bases de datos ni frameworks.
type ProductService struct {
	repo   domain.ProductRepository
	logger *zap.Logger
}

func NewProductService(repo domain.ProductRepository, logger *zap.Logger) *ProductService {
	return &ProductService{
		repo:   repo,
		logger: logger,
	}
}

// CreateProduct orquesta la creación de un product.
func (s *ProductService) CreateProduct(ctx context.Context, sku string, name string, price float64) (*domain.Product, error) {
	e, err := domain.NewProduct(sku, name, price)
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateIfNotExists(ctx, e); err != nil {
		if errors.Is(err, domain.ErrProductDuplicated) {
			return nil, err
		}
		s.logger.Error("failed to create product", zap.Error(err))
		return nil, fmt.Errorf("creating product: %w", err)
	}

	s.logger.Info("product created", zap.String("id", e.ID.String()))
	return e, nil
}

// GetByID busca un product por su ID.
func (s *ProductService) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}
	e, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding product: %w", err)
	}
	return e, nil
}

// ListProducts retorna todos los registros.
func (s *ProductService) ListProducts(ctx context.Context) ([]*domain.Product, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing products: %w", err)
	}
	return items, nil
}
