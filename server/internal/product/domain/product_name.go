package domain

import "unicode/utf8"

// ProductName は商品名のバリューオブジェクトです。
type ProductName struct {
	value string
}

// NewProductName は商品名を生成します。
func NewProductName(v string) (ProductName, error) {
	n := utf8.RuneCountInString(v)
	if n < 1 || n > 20 {
		return ProductName{}, ErrInvalidProductName
	}
	return ProductName{value: v}, nil
}

// Value は商品名の値を返します。
func (n ProductName) Value() string {
	return n.value
}
