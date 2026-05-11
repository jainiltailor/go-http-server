package models

import "time"

// ──────────────────────────────────────────────────────────────────────────────
// API envelope types
// ──────────────────────────────────────────────────────────────────────────────

// APIResponse is the standard JSON envelope for all successful responses.
//
//	{
//	  "status":  "success",
//	  "data":    { ... },
//	  "meta":    { ... }
//	}
type APIResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Meta   *Meta       `json:"meta,omitempty"`
}

// APIError is the standard JSON envelope for all error responses.
//
//	{
//	  "status":  "error",
//	  "code":    "NOT_FOUND",
//	  "message": "Product not found"
//	}
type APIError struct {
	Status  string `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Meta holds pagination and request tracing information.
type Meta struct {
	RequestID  string `json:"request_id,omitempty"`
	Page       int    `json:"page,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	TotalCount int    `json:"total_count,omitempty"`
}

// ──────────────────────────────────────────────────────────────────────────────
// Domain models
// ──────────────────────────────────────────────────────────────────────────────

// Product represents a product in our e-commerce example domain.
type Product struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateProductRequest is the body expected when creating a product.
type CreateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	Category    string  `json:"category"`
}

// HealthResponse is returned by the /health endpoint.
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}
