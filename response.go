package utils

import (
	"encoding/json"
	"net/http"

	"github.com/engineermentor/go-http-server/internal/models"
)

// WriteJSON serializes v to JSON and writes it to w with the given HTTP status.
// Always sets Content-Type: application/json.
func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		// At this point headers are already sent; best we can do is log.
		http.Error(w, "JSON encoding failed", http.StatusInternalServerError)
	}
}

// WriteSuccess wraps data in the standard success envelope.
func WriteSuccess(w http.ResponseWriter, status int, data interface{}, meta *models.Meta) {
	WriteJSON(w, status, models.APIResponse{
		Status: "success",
		Data:   data,
		Meta:   meta,
	})
}

// WriteError writes a structured error response.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, models.APIError{
		Status:  "error",
		Code:    code,
		Message: message,
	})
}

// GetRequestID retrieves the request-id injected by middleware.
func GetRequestID(r *http.Request) string {
	if id, ok := r.Context().Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// RequestIDKey is the context key for the request ID.
type contextKey string

const RequestIDKey contextKey = "request_id"
