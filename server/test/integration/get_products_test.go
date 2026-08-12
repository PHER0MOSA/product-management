package integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetProducts(t *testing.T) {
	router := setupRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var got []productJSON
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) < 3 {
		t.Fatalf("len(got) = %d, want at least 3", len(got))
	}

	byID := map[int64]productJSON{}
	for _, p := range got {
		byID[p.ID] = p
	}
	want := productJSON{ID: 1, Name: "商品A", Price: 1000}
	if byID[1] != want {
		t.Errorf("product id=1 = %+v, want %+v", byID[1], want)
	}
}
