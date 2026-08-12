package application

import (
	"context"
	"fmt"

	"products-manage-server/internal/product/domain"
)

// ProductService は商品に関するユースケースを提供する。
type ProductService struct {
	repo domain.ProductRepository
}

// NewProductService は ProductService を生成する。
func NewProductService(repo domain.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

// ListProducts は商品一覧を返す。
func (s *ProductService) ListProducts(ctx context.Context) ([]*domain.Product, error) {
	products, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("[application] list products: %w", err)
	}
	return products, nil
}

// GetProductByID は ID で商品を1件返す。見つからない場合は domain.ErrProductNotFound。
func (s *ProductService) GetProductByID(ctx context.Context, id domain.ProductID) (*domain.Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("[application] get product by id: %w", err)
	}
	return product, nil
}
