package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/engineermentor/go-http-server/internal/models"
	"github.com/engineermentor/go-http-server/internal/services"
	"github.com/engineermentor/go-http-server/internal/utils"
)

// ──────────────────────────────────────────────────────────────────────────────
// Router  — central route registration
// ──────────────────────────────────────────────────────────────────────────────

// RegisterRoutes wires all application routes onto mux.
// Pattern:  METHOD /path  →  handler
//
// Go 1.22+ ServeMux supports method prefixes and path parameters
// (e.g. "GET /products/{id}") natively — no third-party router needed
// for typical REST APIs.
func RegisterRoutes(mux *http.ServeMux) {
	productSvc := services.NewProductService()
	ph := &ProductHandler{svc: productSvc}

	// Health / readiness
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /ready", readyHandler)

	// API v1 — Products (full CRUD)
	mux.HandleFunc("GET /api/v1/products", ph.List)
	mux.HandleFunc("POST /api/v1/products", ph.Create)
	mux.HandleFunc("GET /api/v1/products/{id}", ph.GetByID)
	mux.HandleFunc("DELETE /api/v1/products/{id}", ph.Delete)

	// Catch-all: return structured 404
	mux.HandleFunc("/", notFoundHandler)
}

// ──────────────────────────────────────────────────────────────────────────────
// Health handlers
// ──────────────────────────────────────────────────────────────────────────────

func healthHandler(w http.ResponseWriter, r *http.Request) {
	utils.WriteSuccess(w, http.StatusOK, models.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Version:   "1.0.0",
	}, nil)
}

// readyHandler would check DB connectivity, cache, external deps, etc.
func readyHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate readiness check
	utils.WriteSuccess(w, http.StatusOK, map[string]string{
		"database": "ok",
		"cache":    "ok",
	}, nil)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	utils.WriteError(w, http.StatusNotFound, "NOT_FOUND",
		"The requested endpoint does not exist")
}

// ──────────────────────────────────────────────────────────────────────────────
// ProductHandler — all product-related HTTP handlers
// ──────────────────────────────────────────────────────────────────────────────

// ProductHandler groups handlers that share the product service dependency.
// This is the "handler struct" pattern — much cleaner than passing deps via
// package-level globals.
type ProductHandler struct {
	svc *services.ProductService
}

// List  GET /api/v1/products
func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	products := h.svc.GetAll()

	// Optional: filter by category query param  ?category=electronics
	if cat := r.URL.Query().Get("category"); cat != "" {
		var filtered []*models.Product
		for _, p := range products {
			if strings.EqualFold(p.Category, cat) {
				filtered = append(filtered, p)
			}
		}
		products = filtered
	}

	utils.WriteSuccess(w, http.StatusOK, products, &models.Meta{
		RequestID:  utils.GetRequestID(r),
		TotalCount: len(products),
	})
}

// GetByID  GET /api/v1/products/{id}
func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// Go 1.22: r.PathValue("id") extracts the {id} segment from the URL
	id := r.PathValue("id")
	if id == "" {
		utils.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "id is required")
		return
	}

	product, err := h.svc.GetByID(id)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, product, &models.Meta{
		RequestID: utils.GetRequestID(r),
	})
}

// Create  POST /api/v1/products
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	// 1. Decode body — always limit request body size in production!
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB

	var req models.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "Request body is not valid JSON: "+err.Error())
		return
	}

	// 2. Delegate to service (business logic lives there, not here)
	product, err := h.svc.Create(req)
	if err != nil {
		utils.WriteError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}

	// 3. Return 201 Created with the new resource
	w.Header().Set("Location", "/api/v1/products/"+product.ID)
	utils.WriteSuccess(w, http.StatusCreated, product, &models.Meta{
		RequestID: utils.GetRequestID(r),
	})
}

// Delete  DELETE /api/v1/products/{id}
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "id is required")
		return
	}

	if err := h.svc.Delete(id); err != nil {
		utils.WriteError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}

	// 204 No Content — success with no body
	w.WriteHeader(http.StatusNoContent)
}
