package handler

import (
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
func writeError(w http.ResponseWriter, status int, message string) {
	if status == http.StatusInternalServerError {
		log.Printf("Internal Server Error: %v", message)
		message = "Internal Server Error"
	}

	writeJSON(w, status, domain.ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}
