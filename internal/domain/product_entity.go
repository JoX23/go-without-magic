package domain

import (
	"time"

	"github.com/google/uuid"
)

// ProductStatus representa los valores posibles del campo Status.
type ProductStatus string

const (
	ProductStatusDraft     ProductStatus = "draft"
	ProductStatusPublished ProductStatus = "published"
	ProductStatusArchived  ProductStatus = "archived"
)

// Product es la entidad central del dominio.
// NO depende de ningún framework, ORM ni capa de transporte.
type Product struct {
	ID          uuid.UUID
	Sku         string
	Name        string
	Price       float64
	Description *string
	Status      ProductStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewProduct es el único constructor válido.
// Garantiza que la entidad siempre nace en estado válido.
func NewProduct(sku string, name string, price float64) (*Product, error) {
	if sku == "" {
		return nil, ErrInvalidProductSku
	}
	if name == "" {
		return nil, ErrInvalidProductName
	}
	if price == 0 {
		return nil, ErrInvalidProductPrice
	}

	now := time.Now().UTC()

	return &Product{
		ID:          uuid.New(),
		Sku:         sku,
		Name:        name,
		Price:       price,
		Description: nil,
		Status:      ProductStatus("draft"),
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}
