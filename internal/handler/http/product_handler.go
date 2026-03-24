package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/JoX23/go-without-magic/internal/domain"
	"github.com/JoX23/go-without-magic/internal/service"
)

// ProductHandler maneja las peticiones HTTP para el recurso Product.
// Responsabilidad única: traducir HTTP ↔ servicio de dominio.
type ProductHandler struct {
	svc    *service.ProductService
	logger *zap.Logger
}

func NewProductHandler(svc *service.ProductService, logger *zap.Logger) *ProductHandler {
	return &ProductHandler{svc: svc, logger: logger}
}

// RegisterRoutes registra todas las rutas de Product en el mux dado.
func (h *ProductHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /products", h.CreateProduct)
	mux.HandleFunc("GET /products", h.ListProducts)
	mux.HandleFunc("GET /products/{id}", h.GetProduct)
}

// CreateProduct maneja POST /products
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Sku   string  `json:"sku"`
		Name  string  `json:"name"`
		Price float64 `json:"price"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProductError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	e, err := h.svc.CreateProduct(r.Context(), req.Sku, req.Name, req.Price)
	if err != nil {
		h.handleProductError(w, err)
		return
	}

	writeProductJSON(w, http.StatusCreated, toProductResponse(e))
}

// GetProduct maneja GET /products/{id}
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	e, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		h.handleProductError(w, err)
		return
	}

	writeProductJSON(w, http.StatusOK, toProductResponse(e))
}

// ListProducts maneja GET /products
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListProducts(r.Context())
	if err != nil {
		h.handleProductError(w, err)
		return
	}

	resp := make([]productResponse, 0, len(items))
	for _, e := range items {
		resp = append(resp, toProductResponse(e))
	}

	writeProductJSON(w, http.StatusOK, resp)
}

// ── Helpers ────────────────────────────────────────────────────────────────

type productResponse struct {
	ID          string  `json:"id"`
	Sku         string  `json:"sku"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description *string `json:"description,omitempty"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
}

func toProductResponse(e *domain.Product) productResponse {
	return productResponse{
		ID:          e.ID.String(),
		Sku:         e.Sku,
		Name:        e.Name,
		Price:       e.Price,
		Description: e.Description,
		Status:      string(e.Status),
		CreatedAt:   e.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (h *ProductHandler) handleProductError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrProductNotFound):
		writeProductError(w, http.StatusNotFound, "product not found")
	case errors.Is(err, domain.ErrProductDuplicated):
		writeProductError(w, http.StatusConflict, "product already exists")
	case errors.Is(err, domain.ErrInvalidProductSku):
		writeProductError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrInvalidProductName):
		writeProductError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrInvalidProductPrice):
		writeProductError(w, http.StatusBadRequest, err.Error())
	default:
		h.logger.Error("unhandled error", zap.Error(err))
		writeProductError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeProductJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body) //nolint:errcheck
}

func writeProductError(w http.ResponseWriter, status int, msg string) {
	writeProductJSON(w, status, map[string]string{"error": msg})
}
