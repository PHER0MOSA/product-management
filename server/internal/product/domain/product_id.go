package domain

// ProductID は商品IDのバリューオブジェクトです。
type ProductID struct {
	value int64
}

// NewProductID は商品IDを生成します。
func NewProductID(v int64) (ProductID, error) {
	if v < 1 {
		return ProductID{}, ErrInvalidProductID
	}
	return ProductID{value: v}, nil
}

// Value は商品IDの値を返します。
func (id ProductID) Value() int64 {
	return id.value
}
