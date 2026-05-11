package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/engineermentor/go-http-server/internal/models"
)

// ProductService handles all product business logic.
// In a real system this would call a repository that talks to Postgres/Redis.
// We use an in-memory map here to keep the focus on HTTP + routing patterns.
type ProductService struct {
	mu       sync.RWMutex           // Protect concurrent map access
	products map[string]*models.Product
	counter  int
}

// NewProductService creates the service and seeds a few demo products.
func NewProductService() *ProductService {
	svc := &ProductService{
		products: make(map[string]*models.Product),
	}
	svc.seed()
	return svc
}

// GetAll returns every product (no pagination for simplicity).
func (s *ProductService) GetAll() []*models.Product {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*models.Product, 0, len(s.products))
	for _, p := range s.products {
		result = append(result, p)
	}
	return result
}

// GetByID returns a single product or an error if not found.
func (s *ProductService) GetByID(id string) (*models.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.products[id]
	if !ok {
		return nil, fmt.Errorf("product %q not found", id)
	}
	return p, nil
}

// Create validates and persists a new product.
func (s *ProductService) Create(req models.CreateProductRequest) (*models.Product, error) {
	// Basic validation — in production use a validation library
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.Price < 0 {
		return nil, fmt.Errorf("price must be non-negative")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter++
	id := fmt.Sprintf("prod_%04d", s.counter)
	now := time.Now().UTC()

	p := &models.Product{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Category:    req.Category,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.products[id] = p
	return p, nil
}

// Delete removes a product by ID.
func (s *ProductService) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.products[id]; !ok {
		return fmt.Errorf("product %q not found", id)
	}
	delete(s.products, id)
	return nil
}

// seed populates the store with demo data.
func (s *ProductService) seed() {
	demos := []models.CreateProductRequest{
		{Name: "Go Programming Book", Description: "Master Go from zero to production", Price: 49.99, Stock: 100, Category: "books"},
		{Name: "Mechanical Keyboard", Description: "Cherry MX Red switches, RGB", Price: 129.00, Stock: 50, Category: "electronics"},
		{Name: "Standing Desk", Description: "Adjustable height, 140cm wide", Price: 399.00, Stock: 20, Category: "furniture"},
	}
	for _, d := range demos {
		_, _ = s.Create(d)
	}
}
