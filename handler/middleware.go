package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
)

// contextKey is an unexported type for context keys in this package,
// preventing collisions with keys defined in other packages.
type contextKey string

const requestIDKey contextKey = "id"

// requestID retrieves the request ID stored in the context by the middleware.
// Returns an empty string if no ID is present.
func requestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// RequestIDMiddleware injects a UUID into the request context under the key "id"
// and logs the incoming request method and path with that ID.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		log.Printf("[%s] %s %s", id, r.Method, r.URL.Path)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
