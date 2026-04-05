package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"presto_tou_service/domain"
)

type HttpHandler struct {
	service domain.PricingService
}

func NewHttpHandler(service domain.PricingService) *HttpHandler {
	return &HttpHandler{service: service}
}

// writeJSON writes a JSON response with the given status code and payload.
func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

// writeError writes a JSON error response with the given status code and message.
// The request ID from ctx is included in the log for internal server errors.
func writeError(ctx context.Context, w http.ResponseWriter, status int, message string) {
	if status == http.StatusInternalServerError {
		log.Printf("[%s] Internal Server Error: %v", requestID(ctx), message)
		message = "Internal Server Error"
	}

	writeJSON(w, status, domain.ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}
