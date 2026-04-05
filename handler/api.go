package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"presto_tou_service/constants"
	"presto_tou_service/domain"
	"presto_tou_service/utils"
	"strings"
	"time"
)

// HandleGetPrice godoc
// @Summary      Get price for a charger at a specific time
// @Description  Get the price for a specific charger at a given timestamp
// @Tags         chargers
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Charger ID"
// @Param        timestamp  query     string  true  "Timestamp in ISO8601 format (e.g., 2024-01-15T10:30:00Z)"
// @Success      200  {object}  domain.PriceResponse
// @Failure      400  {object}  domain.ErrorResponse
// @Failure      404  {object}  domain.ErrorResponse
// @Failure      500  {object}  domain.ErrorResponse
// @Router       /chargers/{id}/price [get]
func (h *HttpHandler) HandleGetPrice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chargerID := r.PathValue("id")

	timeParam := strings.TrimSpace(r.URL.Query().Get("timestamp"))
	if timeParam == constants.Empty {
		log.Printf("[%s] HandleGetPrice: missing timestamp query parameter", requestID(ctx))
		writeError(ctx, w, http.StatusBadRequest, "timestamp query parameter is required")
		return
	}

	parsedTime, err := time.Parse(time.RFC3339, timeParam)
	if err != nil {
		log.Printf("[%s] HandleGetPrice: invalid timestamp %q: %v", requestID(ctx), timeParam, err)
		writeError(ctx, w, http.StatusBadRequest, "invalid timestamp format, use ISO8601 (e.g., 2024-01-15T10:30:00Z)")
		return
	}

	resp, err := h.service.GetPriceForTime(ctx, chargerID, parsedTime)
	if err != nil {
		log.Printf("[%s] HandleGetPrice: service error for chargerID=%s: %v", requestID(ctx), chargerID, err)
		status := utils.HttpStatusForError(err)
		writeError(ctx, w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// HandleGetSchedules godoc
// @Summary      Get schedules for a charger
// @Description  Get all schedules for a specific charger
// @Tags         chargers
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Charger ID"
// @Success      200  {object}  domain.SchedulesResponse
// @Failure      400  {object}  domain.ErrorResponse
// @Failure      404  {object}  domain.ErrorResponse
// @Failure      500  {object}  domain.ErrorResponse
// @Router       /chargers/{id}/schedules [get]
func (h *HttpHandler) HandleGetSchedules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chargerID := r.PathValue("id")

	schedules, err := h.service.GetSchedules(ctx, chargerID)
	if err != nil {
		log.Printf("[%s] HandleGetSchedules: service error for chargerID=%s: %v", requestID(ctx), chargerID, err)
		status := utils.HttpStatusForError(err)
		writeError(ctx, w, status, err.Error())
		return
	}

	if len(schedules) == 0 {
		writeJSON(w, http.StatusOK, domain.SchedulesResponse{
			Schedules: []domain.TOUSchedule{},
			Count:     0,
		})
		return
	}

	writeJSON(w, http.StatusOK, domain.SchedulesResponse{
		Schedules: schedules,
		Count:     len(schedules),
	})
}

// HandlePutSchedules godoc
// @Summary      Update all schedules for a charger
// @Description  Replace all schedules for a specific charger
// @Tags         chargers
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Charger ID"
// @Param        schedules  body      []domain.TOUSchedule true "Array of schedules"
// @Success      200  {object}  domain.SuccessResponse
// @Failure      400  {object}  domain.ErrorResponse
// @Failure      500  {object}  domain.ErrorResponse
// @Router       /chargers/{id}/schedules [put]
func (h *HttpHandler) HandlePutSchedules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chargerID := r.PathValue("id")

	var schedules []domain.TOUSchedule
	if err := json.NewDecoder(r.Body).Decode(&schedules); err != nil {
		log.Printf("[%s] HandlePutSchedules: failed to decode request body for chargerID=%s: %v", requestID(ctx), chargerID, err)
		writeError(ctx, w, http.StatusBadRequest, "invalid request body, expected JSON array of schedules")
		return
	}

	if err := h.service.UpdateSchedules(ctx, chargerID, schedules); err != nil {
		log.Printf("[%s] HandlePutSchedules: service error for chargerID=%s: %v", requestID(ctx), chargerID, err)
		status := utils.HttpStatusForError(err)
		writeError(ctx, w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domain.SuccessResponse{Message: "schedules updated successfully"})
}

// HandlePatchSchedule godoc
// @Summary      Patch a schedule for a charger
// @Description  Update a partial schedule for a specific charger
// @Tags         chargers
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Charger ID"
// @Param        schedule   body      domain.TOUSchedule true "Schedule object"
// @Success      200  {object}  domain.SuccessResponse
// @Failure      400  {object}  domain.ErrorResponse
// @Failure      500  {object}  domain.ErrorResponse
// @Router       /chargers/{id}/schedules [patch]
func (h *HttpHandler) HandlePatchSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chargerID := r.PathValue("id")

	var schedule domain.TOUSchedule
	if err := json.NewDecoder(r.Body).Decode(&schedule); err != nil {
		log.Printf("[%s] HandlePatchSchedule: failed to decode request body for chargerID=%s: %v", requestID(ctx), chargerID, err)
		writeError(ctx, w, http.StatusBadRequest, "invalid request body, expected single schedule object")
		return
	}

	if err := h.service.UpdatePartialSchedule(ctx, chargerID, schedule); err != nil {
		log.Printf("[%s] HandlePatchSchedule: service error for chargerID=%s: %v", requestID(ctx), chargerID, err)
		status := utils.HttpStatusForError(err)
		writeError(ctx, w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domain.SuccessResponse{Message: "schedule updated successfully"})
}

// HandleBulkUpdateSchedules godoc
// @Summary      Bulk update schedules
// @Description  Update schedules for multiple chargers
// @Tags         chargers
// @Accept       json
// @Produce      json
// @Param        request   body      domain.BulkUpdateRequest true "Bulk update request"
// @Success      200  {object}  domain.SuccessResponse
// @Failure      400  {object}  domain.ErrorResponse
// @Failure      500  {object}  domain.ErrorResponse
// @Router       /chargers/bulk/schedules [post]
func (h *HttpHandler) HandleBulkUpdateSchedules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req domain.BulkUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[%s] HandleBulkUpdateSchedules: failed to decode request body: %v", requestID(ctx), err)
		writeError(ctx, w, http.StatusBadRequest, "invalid request body, expected BulkUpdateRequest object")
		return
	}

	if err := h.service.BulkUpdateSchedules(ctx, req.ChargerIDs, req.Schedules); err != nil {
		log.Printf("[%s] HandleBulkUpdateSchedules: service error for chargerIDs=%v: %v", requestID(ctx), req.ChargerIDs, err)
		status := utils.HttpStatusForError(err)
		writeError(ctx, w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domain.SuccessResponse{Message: "bulk schedules updated successfully"})
}
