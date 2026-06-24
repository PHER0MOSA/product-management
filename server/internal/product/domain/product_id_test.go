package domain

import "testing"

func TestNewProductID(t *testing.T) {
	tests := []struct {
		name    string
		value   int64
		wantErr bool
	}{
		{name: "有効なID（1）", value: 1, wantErr: false},
		{name: "有効なID（大きな正の整数）", value: 999999, wantErr: false},
		{name: "無効なID（0）", value: 0, wantErr: true},
		{name: "無効なID（負の数）", value: -1, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewProductID(tt.value)
			if tt.wantErr && err == nil {
				t.Fatal("error が欲しいが nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("error は欲しくない: %v", err)
			}
		})
	}
}
