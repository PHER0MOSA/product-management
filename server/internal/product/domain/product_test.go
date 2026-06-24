package domain

import "testing"

func TestNewProduct(t *testing.T) {
	validID, err := NewProductID(1)
	if err != nil {
		t.Fatalf("テスト準備に失敗: %v", err)
	}
	validName, err := NewProductName("商品A")
	if err != nil {
		t.Fatalf("テスト準備に失敗: %v", err)
	}
	validPrice, err := NewProductPrice(1000)
	if err != nil {
		t.Fatalf("テスト準備に失敗: %v", err)
	}

	product, err := NewProduct(validID, validName, validPrice)
	if err != nil {
		t.Fatalf("error は欲しくない: %v", err)
	}
	if product == nil {
		t.Fatal("product が nil")
	}
	if product.ID().Value() != 1 {
		t.Fatalf("ID = %d, want 1", product.ID().Value())
	}
}
