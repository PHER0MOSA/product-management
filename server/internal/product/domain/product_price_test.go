package domain

import "testing"

func TestNewProductPrice(t *testing.T) {
	tests := []struct {
		name    string
		value   int64
		wantErr bool
	}{
		{name: "有効な価格（0）", value: 0, wantErr: false},
		{name: "有効な価格（1000）", value: 1000, wantErr: false},
		{name: "無効な価格（負の数）", value: -1, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewProductPrice(tt.value)
			if tt.wantErr && err == nil {
				t.Fatal("error が欲しいが nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("error は欲しくない: %v", err)
			}
		})
	}
}
