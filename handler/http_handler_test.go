package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"presto_tou_service/constants"
	"presto_tou_service/domain"
)

// mockPricingService is a mock implementation of the domain.PricingService interface for testing.
type mockPricingService struct {
	priceResponse *domain.PriceResponse
	priceErr      error
	updateErr     error
	patchErr      error
	bulkUpdateErr error
	schedules     []domain.TOUSchedule
	schedulesErr  error
}

func (m *mockPricingService) GetPriceForTime(ctx context.Context, chargerID string, timestamp time.Time) (*domain.PriceResponse, error) {
	return m.priceResponse, m.priceErr
}

func (m *mockPricingService) UpdateSchedules(ctx context.Context, chargerID string, schedules []domain.TOUSchedule) error {
	return m.updateErr
}

func (m *mockPricingService) UpdatePartialSchedule(ctx context.Context, chargerID string, schedule domain.TOUSchedule) error {
	return m.patchErr
}

func (m *mockPricingService) BulkUpdateSchedules(ctx context.Context, chargerIDs []string, schedules []domain.TOUSchedule) error {
	return m.bulkUpdateErr
}

func (m *mockPricingService) GetSchedules(ctx context.Context, chargerID string) ([]domain.TOUSchedule, error) {
	return m.schedules, m.schedulesErr
}

func newMockService() *mockPricingService {
	return &mockPricingService{
		priceResponse: &domain.PriceResponse{
			ChargerID:        "charger-1",
			RequestedTime:    "2024-01-15T10:30:00Z",
			LocalChargerTime: "2024-01-15T05:30:00-05:00",
			PricingPeriod: domain.PricingPeriod{
				StartTime: "06:00",
				EndTime:   "12:00",
			},
			PricePerKwh: 0.20,
		},
		schedules: []domain.TOUSchedule{
			{StartTime: "00:00", EndTime: "06:00", PricePerKwh: 0.15},
			{StartTime: "06:00", EndTime: "12:00", PricePerKwh: 0.20},
			{StartTime: "12:00", EndTime: "18:00", PricePerKwh: 0.25},
			{StartTime: "18:00", EndTime: "00:00", PricePerKwh: 0.15},
		},
	}
}

func TestHandleGetPrice_Success(t *testing.T) {
	mockSvc := newMockService()
	handler := NewHttpHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/chargers/charger-1/price?timestamp=2024-01-15T10:30:00Z", nil)
	req.SetPathValue("id", "charger-1")
	w := httptest.NewRecorder()

	handler.HandleGetPrice(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var body domain.PriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.ChargerID != "charger-1" {
		t.Errorf("expected charger_id 'charger-1', got '%s'", body.ChargerID)
	}
	if body.PricingPeriod.StartTime != "06:00" {
		t.Errorf("expected start_time '06:00', got '%s'", body.PricingPeriod.StartTime)
	}
	if body.PricingPeriod.EndTime != "12:00" {
		t.Errorf("expected end_time '12:00', got '%s'", body.PricingPeriod.EndTime)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
	}
}

func TestHandleGetPrice_MissingTimestamp(t *testing.T) {
	mockSvc := newMockService()
	handler := NewHttpHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/chargers/charger-1/price", nil)
	req.SetPathValue("id", "charger-1")
	w := httptest.NewRecorder()

	handler.HandleGetPrice(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestHandleGetPrice_InvalidTimestamp(t *testing.T) {
	mockSvc := newMockService()
	handler := NewHttpHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/chargers/charger-1/price?timestamp=not-a-date", nil)
	req.SetPathValue("id", "charger-1")
	w := httptest.NewRecorder()

	handler.HandleGetPrice(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestHandleGetPrice_LeadingSpaceTimestamp(t *testing.T) {
	mockSvc := newMockService()
	handler := NewHttpHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/chargers/charger-1/price?timestamp=%202024-01-15T10:30:00Z", nil)
	req.SetPathValue("id", "charger-1")
	w := httptest.NewRecorder()

	handler.HandleGetPrice(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestHandleGetPrice_ChargerNotFound(t *testing.T) {
	mockSvc := newMockService()
	mockSvc.priceErr = constants.ErrChargerNotFound
	handler := NewHttpHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/chargers/nonexistent/price?timestamp=2024-01-15T10:30:00Z", nil)
	req.SetPathValue("id", "nonexistent")
	w := httptest.NewRecorder()

	handler.HandleGetPrice(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestHandleGetSchedules_Success(t *testing.T) {
	mockSvc := newMockService()
	handler := NewHttpHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/chargers/charger-1/schedules", nil)
	req.SetPathValue("id", "charger-1")
	w := httptest.NewRecorder()

	handler.HandleGetSchedules(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var body domain.SchedulesResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(body.Schedules) != 4 {
		t.Errorf("expected 4 schedules, got %d", len(body.Schedules))
	}
}

func TestHandlePutSchedules_Success(t *testing.T) {
	mockSvc := newMockService()
	handler := NewHttpHandler(mockSvc)

	schedules := []domain.TOUSchedule{
		{StartTime: "00:00", EndTime: "06:00", PricePerKwh: 0.15},
		{StartTime: "06:00", EndTime: "12:00", PricePerKwh: 0.20},
		{StartTime: "12:00", EndTime: "18:00", PricePerKwh: 0.25},
		{StartTime: "18:00", EndTime: "00:00", PricePerKwh: 0.15},
	}
	body, _ := json.Marshal(schedules)

	req := httptest.NewRequest(http.MethodPut, "/chargers/charger-1/schedules", bytes.NewReader(body))
	req.SetPathValue("id", "charger-1")
	w := httptest.NewRecorder()

	handler.HandlePutSchedules(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestHandlePutSchedules_InvalidBody(t *testing.T) {
	mockSvc := newMockService()
	handler := NewHttpHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPut, "/chargers/charger-1/schedules", bytes.NewReader([]byte("invalid json")))
	req.SetPathValue("id", "charger-1")
	w := httptest.NewRecorder()

	handler.HandlePutSchedules(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestHandlePutSchedules_ValidationError(t *testing.T) {
	mockSvc := newMockService()
	mockSvc.updateErr = constants.ErrIncompleteDayCoverage
	handler := NewHttpHandler(mockSvc)

	schedules := []domain.TOUSchedule{
		{StartTime: "00:00", EndTime: "06:00", PricePerKwh: 0.15},
	}
	body, _ := json.Marshal(schedules)

	req := httptest.NewRequest(http.MethodPut, "/chargers/charger-1/schedules", bytes.NewReader(body))
	req.SetPathValue("id", "charger-1")
	w := httptest.NewRecorder()

	handler.HandlePutSchedules(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestHandlePatchSchedule_Success(t *testing.T) {
	mockSvc := newMockService()
	handler := NewHttpHandler(mockSvc)

	schedule := domain.TOUSchedule{
		StartTime:   "06:00",
		EndTime:     "12:00",
		PricePerKwh: 0.22,
	}
	body, _ := json.Marshal(schedule)

	req := httptest.NewRequest(http.MethodPatch, "/chargers/charger-1/schedules", bytes.NewReader(body))
	req.SetPathValue("id", "charger-1")
	w := httptest.NewRecorder()

	handler.HandlePatchSchedule(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestHandlePatchSchedule_InvalidBody(t *testing.T) {
	mockSvc := newMockService()
	handler := NewHttpHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPatch, "/chargers/charger-1/schedules", bytes.NewReader([]byte("invalid json")))
	req.SetPathValue("id", "charger-1")
	w := httptest.NewRecorder()

	handler.HandlePatchSchedule(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestHandleBulkUpdateSchedules_Success(t *testing.T) {
	mockSvc := newMockService()
	handler := NewHttpHandler(mockSvc)

	reqBody := domain.BulkUpdateRequest{
		ChargerIDs: []string{"charger-1", "charger-2"},
		Schedules: []domain.TOUSchedule{
			{StartTime: "00:00", EndTime: "06:00", PricePerKwh: 0.15},
			{StartTime: "06:00", EndTime: "12:00", PricePerKwh: 0.20},
			{StartTime: "12:00", EndTime: "18:00", PricePerKwh: 0.25},
			{StartTime: "18:00", EndTime: "00:00", PricePerKwh: 0.15},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/chargers/bulk/schedules", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.HandleBulkUpdateSchedules(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestHandleBulkUpdateSchedules_InvalidBody(t *testing.T) {
	mockSvc := newMockService()
	handler := NewHttpHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/chargers/bulk/schedules", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.HandleBulkUpdateSchedules(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	payload := map[string]string{"key": "value"}

	writeJSON(w, http.StatusOK, payload)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["key"] != "value" {
		t.Errorf("expected 'value', got '%s'", body["key"])
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()

	writeError(w, http.StatusBadRequest, "test error message")

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	var body domain.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.Error != "Bad Request" {
		t.Errorf("expected error 'Bad Request', got '%s'", body.Error)
	}
	if body.Message != "test error message" {
		t.Errorf("expected message 'test error message', got '%s'", body.Message)
	}
}
