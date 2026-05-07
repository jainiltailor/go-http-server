package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/engineermentor/go-http-server/internal/utils"
)

// MiddlewareFunc is a standard middleware type: wraps an http.Handler.
type MiddlewareFunc func(http.Handler) http.Handler

// Chain applies middlewares from right to left so the first middleware
// in the list is the outermost (first to execute on request).
//
//	Chain(mux, A, B, C) → A(B(C(mux)))
//	Request flow: A → B → C → mux → C → B → A
func Chain(h http.Handler, middlewares ...MiddlewareFunc) http.Handler {
	// Apply in reverse so that the first arg is the outermost wrapper
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// ──────────────────────────────────────────────────────────────────────────────
// RequestID — inject a unique ID into every request context
// ──────────────────────────────────────────────────────────────────────────────

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Honour an upstream ID (e.g. from a load balancer) if present
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = fmt.Sprintf("%d", time.Now().UnixNano())
		}

		// Store in context so handlers can read it
		ctx := context.WithValue(r.Context(), utils.RequestIDKey, id)
		w.Header().Set("X-Request-ID", id)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// Logger — structured request logging
// ──────────────────────────────────────────────────────────────────────────────

// responseWriter wraps http.ResponseWriter to capture the status code written
// by the handler so we can log it after the response is sent.
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		requestID := utils.GetRequestID(r)

		log.Printf(
			"[%s] method=%s path=%s status=%d size=%dB duration=%s ip=%s",
			requestID,
			r.Method,
			r.URL.Path,
			rw.status,
			rw.size,
			duration,
			r.RemoteAddr,
		)
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// CORS — allow cross-origin requests (configure for production)
// ──────────────────────────────────────────────────────────────────────────────

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// In production, replace * with your actual frontend origin.
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")

		// Pre-flight request — browsers send this before the real request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// Recover — catch panics and return 500 instead of crashing
// ──────────────────────────────────────────────────────────────────────────────

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC recovered: %v\n%s", err, debug.Stack())
				http.Error(w, `{"status":"error","code":"INTERNAL_ERROR","message":"An unexpected error occurred"}`,
					http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
