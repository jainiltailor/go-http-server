package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/engineermentor/go-http-server/configs"
	"github.com/engineermentor/go-http-server/internal/handlers"
	"github.com/engineermentor/go-http-server/internal/middleware"
)

func main() {
	// Load config
	cfg := configs.Load()

	// Build the router
	mux := http.NewServeMux()

	// Register all routes
	handlers.RegisterRoutes(mux)

	// Wrap mux with middleware chain
	// Order: Logger → RequestID → CORS → Recover → mux
	handler := middleware.Chain(
		mux,
		middleware.RequestID,
		middleware.Logger,
		middleware.CORS,
		middleware.Recover,
	)

	// Create production-grade HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine so we can listen for shutdown
	go func() {
		log.Printf("🚀 Server starting on http://localhost:%s\n", cfg.Port)
		log.Printf("📋 Environment: %s\n", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// ── Graceful Shutdown ──────────────────────────────────────────────────
	// Block until OS signal received (Ctrl+C or SIGTERM from k8s)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("⏳ Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Forced shutdown: %v", err)
	}

	log.Println("✅ Server exited cleanly")
}
