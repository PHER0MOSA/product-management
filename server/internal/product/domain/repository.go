package domain

import "context"

// ProductRepository は商品の永続化を抽象化する。
// 実装は外側の infrastructure に置く。
type ProductRepository interface {
	List(ctx context.Context) ([]*Product, error)
	GetByID(ctx context.Context, id ProductID) (*Product, error)
}
