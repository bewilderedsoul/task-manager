// Package httpx contains small helpers for writing consistent JSON responses
// and error envelopes across all handlers.
package httpx

import (
	"encoding/json"
	"net/http"
)

// ErrorBody is the consistent error envelope returned by every endpoint:
//
//	{ "error": { "message": "...", "details": { ... } } }
type ErrorBody struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// JSON writes v as a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// Error writes a consistent error envelope.
func Error(w http.ResponseWriter, status int, message string, details ...any) {
	body := ErrorBody{Error: ErrorDetail{Message: message}}
	if len(details) > 0 {
		body.Error.Details = details[0]
	}
	JSON(w, status, body)
}
