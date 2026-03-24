package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JoX23/go-without-magic/internal/config"
	"github.com/JoX23/go-without-magic/internal/domain"
)

type ProductRepository struct {
	pool *pgxpool.Pool
}

func NewProductRepository(cfg config.DatabaseConfig) (*ProductRepository, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("parsing database DSN: %w", err)
	}

	poolCfg.MaxConns = int32(cfg.MaxOpenConns)
	poolCfg.MinConns = int32(cfg.MaxIdleConns)

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return &ProductRepository{pool: pool}, nil
}

func (r *ProductRepository) CreateIfNotExists(ctx context.Context, e *domain.Product) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO products (id, sku, name, price, description, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		e.ID.String(), e.Sku, e.Name, e.Price, e.Description, e.Status, e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("creating product: %w", err)
	}
	return nil
}

func (r *ProductRepository) Save(ctx context.Context, e *domain.Product) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO products (id, sku, name, price, description, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		e.ID.String(), e.Sku, e.Name, e.Price, e.Description, e.Status, e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting product: %w", err)
	}
	return nil
}

func (r *ProductRepository) FindByID(ctx context.Context, id string) (*domain.Product, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, sku, name, price, description, status, created_at, updated_at
		 FROM products WHERE id = $1`,
		id,
	)
	e, err := scanProduct(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrProductNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying by id: %w", err)
	}
	return e, nil
}

func (r *ProductRepository) FindBySku(ctx context.Context, sku string) (*domain.Product, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, sku, name, price, description, status, created_at, updated_at
		 FROM products WHERE sku = $1`,
		sku,
	)
	e, err := scanProduct(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrProductNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying by sku: %w", err)
	}
	return e, nil
}

func (r *ProductRepository) List(ctx context.Context) ([]*domain.Product, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, sku, name, price, description, status, created_at, updated_at
		 FROM products ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("listing products: %w", err)
	}
	defer rows.Close()

	var items []*domain.Product
	for rows.Next() {
		e, err := scanProduct(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}
		items = append(items, e)
	}
	return items, rows.Err()
}

type productScanner interface {
	Scan(dest ...any) error
}

func scanProduct(s productScanner) (*domain.Product, error) {
	var e domain.Product
	var idStr string

	if err := s.Scan(&idStr, &e.Sku, &e.Name, &e.Price, &e.Description, &e.Status, &e.CreatedAt, &e.UpdatedAt); err != nil {
		return nil, err
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("parsing uuid: %w", err)
	}
	e.ID = id

	return &e, nil
}
