package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"products-manage-server/internal/product/application"
	"products-manage-server/internal/product/domain"
)

// productResponse は openapi の Product に対応する JSON 用 DTO。
type productResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Price int64  `json:"price"`
}

// commonError は openapi の CommonError に対応する JSON 用 DTO。
// エラー種別は HTTP ステータスで表し、body はメッセージのみ。
type commonError struct {
	Message string `json:"message"`
}

// ProductHandler は商品 API の HTTP ハンドラ。
type ProductHandler struct {
	service *application.ProductService
}

// NewProductHandler は ProductHandler を生成する。
func NewProductHandler(service *application.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

// Register は商品関連のルートを登録する。
func (h *ProductHandler) Register(r *mux.Router) {
	r.HandleFunc("/products", h.ListProducts).Methods(http.MethodGet)
	r.HandleFunc("/products/{id}", h.GetProductByID).Methods(http.MethodGet)
}

// ListProducts は GET /products。
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.service.ListProducts(r.Context())
	if err != nil {
		log.Printf("[handler] ListProducts: %v", err)
		writeCommonError(w, http.StatusInternalServerError, MessageInternalError)
		return
	}

	resp := make([]productResponse, 0, len(products))
	for _, p := range products {
		resp = append(resp, toProductResponse(p))
	}
	writeJSON(w, http.StatusOK, resp)
}

// GetProductByID は GET /products/{id}。
func (h *ProductHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	rawID := mux.Vars(r)["id"]
	idNum, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		writeCommonError(w, http.StatusNotFound, MessageProductNotFound)
		return
	}

	id, err := domain.NewProductID(idNum)
	if err != nil {
		writeCommonError(w, http.StatusNotFound, MessageProductNotFound)
		return
	}

	product, err := h.service.GetProductByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			writeCommonError(w, http.StatusNotFound, MessageProductNotFound)
			return
		}
		log.Printf("[handler] GetProductByID: %v", err)
		writeCommonError(w, http.StatusInternalServerError, MessageInternalError)
		return
	}

	writeJSON(w, http.StatusOK, toProductResponse(product))
}

func toProductResponse(p *domain.Product) productResponse {
	return productResponse{
		ID:    p.ID().Value(),
		Name:  p.Name.Value(),
		Price: p.Price.Value(),
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeCommonError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, commonError{Message: message})
}
