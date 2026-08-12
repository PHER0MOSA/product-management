package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"products-manage-server/internal/product/domain"
)

// ProductRepository は PostgreSQL 向けの domain.ProductRepository 実装。
type ProductRepository struct {
	db *sql.DB
}

var _ domain.ProductRepository = (*ProductRepository)(nil)

// NewProductRepository は ProductRepository を生成する。
func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// List は商品をすべて返す。
func (r *ProductRepository) List(ctx context.Context) ([]*domain.Product, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, price FROM products ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, fmt.Errorf("[repository] list products: %w", err)
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	return products, nil
}

// GetByID は ID で商品を1件返す。
func (r *ProductRepository) GetByID(ctx context.Context, id domain.ProductID) (*domain.Product, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, name, price FROM products WHERE id = $1`, id.Value())
	p, err := scanProduct(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("[repository] get product by id: %w", domain.ErrProductNotFound)
		}
		return nil, fmt.Errorf("[repository] get product by id: %w", err)
	}
	return p, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanProduct(s rowScanner) (*domain.Product, error) {
	var (
		id    int64
		name  string
		price int64
	)
	if err := s.Scan(&id, &name, &price); err != nil {
		return nil, err
	}

	productID, err := domain.NewProductID(id)
	if err != nil {
		return nil, fmt.Errorf("[repository] scan product id: %w", err)
	}
	productName, err := domain.NewProductName(name)
	if err != nil {
		return nil, fmt.Errorf("[repository] scan product name: %w", err)
	}
	productPrice, err := domain.NewProductPrice(price)
	if err != nil {
		return nil, fmt.Errorf("[repository] scan product price: %w", err)
	}
	return domain.NewProduct(productID, productName, productPrice)
}
