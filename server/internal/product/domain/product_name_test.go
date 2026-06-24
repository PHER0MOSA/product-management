package domain

import (
	"strings"
	"testing"
)

func TestNewProductName(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "有効な名前", value: "商品A", wantErr: false},
		{name: "有効な名前（20文字ちょうど）", value: strings.Repeat("あ", 20), wantErr: false},
		{name: "無効な名前（空文字）", value: "", wantErr: true},
		{name: "無効な名前（21文字以上）", value: strings.Repeat("あ", 21), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewProductName(tt.value)
			if tt.wantErr && err == nil {
				t.Fatal("error が欲しいが nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("error は欲しくない: %v", err)
			}
		})
	}
}
