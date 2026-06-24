package domain

import "errors"

var (
	ErrInvalidProductID    = errors.New("product id must be >= 1")
	ErrInvalidProductName  = errors.New("product name must be 1-20 characters")
	ErrInvalidProductPrice = errors.New("product price must be >= 0")
)
