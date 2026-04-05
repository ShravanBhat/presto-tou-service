package router

import (
	"net/http"

	"presto_tou_service/handler"

	httpSwagger "github.com/swaggo/http-swagger" // http-swagger middleware
	_ "presto_tou_service/docs"                  // generated swagger docs
)

// NewRouter sets up all the HTTP routes and returns an http.Handler
// wrapped with request-ID middleware.
func NewRouter(httpHandler *handler.HttpHandler) http.Handler {
	mux := http.NewServeMux()

	// Charger routes
	mux.HandleFunc("GET /chargers/{id}/price", httpHandler.HandleGetPrice)
	mux.HandleFunc("GET /chargers/{id}/schedules", httpHandler.HandleGetSchedules)
	mux.HandleFunc("PUT /chargers/{id}/schedules", httpHandler.HandlePutSchedules)
	mux.HandleFunc("PATCH /chargers/{id}/schedules", httpHandler.HandlePatchSchedule)
	mux.HandleFunc("POST /chargers/bulk/schedules", httpHandler.HandleBulkUpdateSchedules)

	// Swagger route
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	return handler.RequestIDMiddleware(mux)
}
