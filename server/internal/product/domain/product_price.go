package domain

// ProductPrice は商品価格のバリューオブジェクトです。
type ProductPrice struct {
	value int64
}

// NewProductPrice は商品価格を生成します。
func NewProductPrice(v int64) (ProductPrice, error) {
	if v < 0 {
		return ProductPrice{}, ErrInvalidProductPrice
	}
	return ProductPrice{value: v}, nil
}

// Value は商品価格の値を返します。
func (p ProductPrice) Value() int64 {
	return p.value
}
