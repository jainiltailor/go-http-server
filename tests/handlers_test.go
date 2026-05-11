package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/engineermentor/go-http-server/internal/handlers"
	"github.com/engineermentor/go-http-server/internal/models"
)

// newTestServer spins up the full router in-memory.
// httptest.NewServer is perfect for integration-style handler tests.
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux)
	return httptest.NewServer(mux)
}

// ──────────────────────────────────────────────────────────────────────────────
// Health endpoint
// ──────────────────────────────────────────────────────────────────────────────

func TestHealthEndpoint(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}

	var body models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Status != "success" {
		t.Errorf("expected status=success, got %s", body.Status)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Products — list (seeded data)
// ──────────────────────────────────────────────────────────────────────────────

func TestListProducts(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/products")
	if err != nil {
		t.Fatalf("GET /api/v1/products failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var body models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if body.Meta == nil || body.Meta.TotalCount < 3 {
		t.Errorf("expected at least 3 seeded products, meta=%+v", body.Meta)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Products — create
// ──────────────────────────────────────────────────────────────────────────────

func TestCreateProduct(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	payload := models.CreateProductRequest{
		Name:     "Test Laptop",
		Price:    999.99,
		Stock:    10,
		Category: "electronics",
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(
		srv.URL+"/api/v1/products",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("POST /api/v1/products failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}

	// Location header must be set
	if loc := resp.Header.Get("Location"); loc == "" {
		t.Error("expected Location header to be set")
	}

	var apiResp models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if apiResp.Status != "success" {
		t.Errorf("expected success, got %s", apiResp.Status)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Products — validation error
// ──────────────────────────────────────────────────────────────────────────────

func TestCreateProduct_MissingName(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	payload := models.CreateProductRequest{Price: 100} // missing name
	body, _ := json.Marshal(payload)

	resp, err := http.Post(
		srv.URL+"/api/v1/products",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", resp.StatusCode)
	}

	var errResp models.APIError
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if errResp.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %s", errResp.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// 404 for unknown routes
// ──────────────────────────────────────────────────────────────────────────────

func TestUnknownRoute(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/unknown")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}
