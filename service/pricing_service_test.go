package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"presto_tou_service/constants"
	"presto_tou_service/domain"
)

// mockRepository is a mock implementation of the domain.Repository interface for testing.
type mockRepository struct {
	charger        *domain.Charger
	chargerErr     error
	period         *domain.PricingPeriod
	periodErr      error
	schedules      []domain.TOUSchedule
	scheduleErr    error
	replaceErr     error
	updateErr      error
	bulkReplaceErr error
}

func (m *mockRepository) GetChargerByID(ctx context.Context, chargerID string) (*domain.Charger, error) {
	return m.charger, m.chargerErr
}

func (m *mockRepository) GetPriceAtTime(ctx context.Context, chargerID string, localTime time.Time) (*domain.PricingPeriod, float64, error) {
	return m.period, 0, m.periodErr
}

func (m *mockRepository) ReplaceSchedules(ctx context.Context, chargerID string, schedules []domain.TOUSchedule) error {
	return m.replaceErr
}

func (m *mockRepository) UpdateSingleSchedule(ctx context.Context, chargerID string, schedule domain.TOUSchedule) error {
	return m.updateErr
}

func (m *mockRepository) GetSchedulesByChargerID(ctx context.Context, chargerID string) ([]domain.TOUSchedule, error) {
	return m.schedules, m.scheduleErr
}

func (m *mockRepository) BulkReplaceSchedules(ctx context.Context, chargerIDs []string, schedules []domain.TOUSchedule) error {
	return m.bulkReplaceErr
}

func newMockRepo() *mockRepository {
	return &mockRepository{
		charger: &domain.Charger{
			ID:       "charger-1",
			Name:     "Test Charger",
			Timezone: "America/New_York",
		},
		period: &domain.PricingPeriod{
			StartTime: "06:00",
			EndTime:   "12:00",
		},
		schedules: []domain.TOUSchedule{
			{StartTime: "00:00", EndTime: "06:00", PricePerKwh: 0.15},
			{StartTime: "06:00", EndTime: "12:00", PricePerKwh: 0.20},
			{StartTime: "12:00", EndTime: "18:00", PricePerKwh: 0.25},
			{StartTime: "18:00", EndTime: "00:00", PricePerKwh: 0.15},
		},
	}
}

func TestGetPriceForTime_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewPricingService(repo)

	utcTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	resp, err := svc.GetPriceForTime(context.Background(), "charger-1", utcTime)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.ChargerID != "charger-1" {
		t.Errorf("expected charger_id 'charger-1', got '%s'", resp.ChargerID)
	}

	if resp.PricingPeriod.StartTime != "06:00" {
		t.Errorf("expected start_time '06:00', got '%s'", resp.PricingPeriod.StartTime)
	}
	if resp.PricingPeriod.EndTime != "12:00" {
		t.Errorf("expected end_time '12:00', got '%s'", resp.PricingPeriod.EndTime)
	}
}

func TestGetPriceForTime_EmptyChargerID(t *testing.T) {
	repo := newMockRepo()
	svc := NewPricingService(repo)

	utcTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	_, err := svc.GetPriceForTime(context.Background(), constants.Empty, utcTime)

	if !errors.Is(err, constants.ErrEmptyChargerID) {
		t.Errorf("expected ErrEmptyChargerID, got %v", err)
	}
}

func TestGetPriceForTime_ChargerNotFound(t *testing.T) {
	repo := newMockRepo()
	repo.chargerErr = constants.ErrChargerNotFound
	svc := NewPricingService(repo)

	utcTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	_, err := svc.GetPriceForTime(context.Background(), "nonexistent", utcTime)

	if !errors.Is(err, constants.ErrChargerNotFound) {
		t.Errorf("expected ErrChargerNotFound, got %v", err)
	}
}

func TestGetPriceForTime_PriceNotFound(t *testing.T) {
	repo := newMockRepo()
	repo.periodErr = constants.ErrPriceNotFound
	svc := NewPricingService(repo)

	utcTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	_, err := svc.GetPriceForTime(context.Background(), "charger-1", utcTime)

	if !errors.Is(err, constants.ErrPriceNotFound) {
		t.Errorf("expected ErrPriceNotFound, got %v", err)
	}
}

func TestUpdateSchedules_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewPricingService(repo)

	// Valid 24-hour coverage with overnight period ending at midnight
	schedules := []domain.TOUSchedule{
		{StartTime: "00:00", EndTime: "06:00", PricePerKwh: 0.15},
		{StartTime: "06:00", EndTime: "12:00", PricePerKwh: 0.20},
		{StartTime: "12:00", EndTime: "18:00", PricePerKwh: 0.25},
		{StartTime: "18:00", EndTime: "00:00", PricePerKwh: 0.15},
	}

	err := svc.UpdateSchedules(context.Background(), "charger-1", schedules)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUpdateSchedules_EmptyChargerID(t *testing.T) {
	repo := newMockRepo()
	svc := NewPricingService(repo)

	schedules := []domain.TOUSchedule{
		{StartTime: "00:00", EndTime: "06:00", PricePerKwh: 0.15},
	}

	err := svc.UpdateSchedules(context.Background(), constants.Empty, schedules)
	if !errors.Is(err, constants.ErrEmptyChargerID) {
		t.Errorf("expected ErrEmptyChargerID, got %v", err)
	}
}

func TestUpdateSchedules_EmptySchedules(t *testing.T) {
	repo := newMockRepo()
	svc := NewPricingService(repo)

	err := svc.UpdateSchedules(context.Background(), "charger-1", []domain.TOUSchedule{})
	if !errors.Is(err, constants.ErrEmptySchedules) {
		t.Errorf("expected ErrEmptySchedules, got %v", err)
	}
}

func TestUpdateSchedules_NegativePrice(t *testing.T) {
	repo := newMockRepo()
	svc := NewPricingService(repo)

	schedules := []domain.TOUSchedule{
		{StartTime: "00:00", EndTime: "06:00", PricePerKwh: -0.10},
		{StartTime: "06:00", EndTime: "12:00", PricePerKwh: 0.20},
		{StartTime: "12:00", EndTime: "18:00", PricePerKwh: 0.25},
		{StartTime: "18:00", EndTime: "00:00", PricePerKwh: 0.15},
	}

	err := svc.UpdateSchedules(context.Background(), "charger-1", schedules)
	if !errors.Is(err, constants.ErrInvalidPrice) {
		t.Errorf("expected ErrInvalidPrice, got %v", err)
	}
}

func TestUpdateSchedules_InvalidTimeFormat(t *testing.T) {
	repo := newMockRepo()
	svc := NewPricingService(repo)

	schedules := []domain.TOUSchedule{
		{StartTime: "25:00", EndTime: "06:00", PricePerKwh: 0.15},
		{StartTime: "06:00", EndTime: "12:00", PricePerKwh: 0.20},
		{StartTime: "12:00", EndTime: "18:00", PricePerKwh: 0.25},
		{StartTime: "18:00", EndTime: "00:00", PricePerKwh: 0.15},
	}

	err := svc.UpdateSchedules(context.Background(), "charger-1", schedules)
	if !errors.Is(err, constants.ErrInvalidTimeFormat) {
		t.Errorf("expected ErrInvalidTimeFormat, got %v", err)
	}
}

func TestUpdateSchedules_IncompleteDayCoverage(t *testing.T) {
	repo := newMockRepo()
	svc := NewPricingService(repo)

	// Only covers 00:00-12:00, missing 12:00-00:00
	schedules := []domain.TOUSchedule{
		{StartTime: "00:00", EndTime: "06:00", PricePerKwh: 0.15},
		{StartTime: "06:00", EndTime: "12:00", PricePerKwh: 0.20},
	}

	err := svc.UpdateSchedules(context.Background(), "charger-1", schedules)
	if !errors.Is(err, constants.ErrIncompleteDayCoverage) {
		t.Errorf("expected ErrIncompleteDayCoverage, got %v", err)
	}
}

func TestUpdateSchedules_OverlappingSchedules(t *testing.T) {
	repo := newMockRepo()
	svc := NewPricingService(repo)

	// 00:00-08:00 and 06:00-12:00 overlap
	schedules := []domain.TOUSchedule{
		{StartTime: "00:00", EndTime: "08:00", PricePerKwh: 0.15},
		{StartTime: "06:00", EndTime: "12:00", PricePerKwh: 0.20},
		{StartTime: "12:00", EndTime: "18:00", PricePerKwh: 0.25},
		{StartTime: "18:00", EndTime: "00:00", PricePerKwh: 0.15},
	}

	err := svc.UpdateSchedules(context.Background(), "charger-1", schedules)
	if !errors.Is(err, constants.ErrScheduleOverlap) {
		t.Errorf("expected ErrScheduleOverlap, got %v", err)
	}
}

func TestUpdatePartialSchedule_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewPricingService(repo)

	schedule := domain.TOUSchedule{
		StartTime:   "06:00",
		EndTime:     "12:00",
		PricePerKwh: 0.22,
	}

	err := svc.UpdatePartialSchedule(context.Background(), "charger-1", schedule)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUpdatePartialSchedule_InvalidTimeFormat(t *testing.T) {
	repo := newMockRepo()
	svc := NewPricingService(repo)

	schedule := domain.TOUSchedule{
		StartTime:   "99:00",
		EndTime:     "99:00",
		PricePerKwh: 0.22,
	}

	err := svc.UpdatePartialSchedule(context.Background(), "charger-1", schedule)
	if !errors.Is(err, constants.ErrInvalidTimeFormat) {
		t.Errorf("expected ErrInvalidTimeFormat, got %v", err)
	}
}

func TestBulkUpdateSchedules_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewPricingService(repo)

	chargerIDs := []string{"charger-1", "charger-2"}
	schedules := []domain.TOUSchedule{
		{StartTime: "00:00", EndTime: "06:00", PricePerKwh: 0.15},
		{StartTime: "06:00", EndTime: "12:00", PricePerKwh: 0.20},
		{StartTime: "12:00", EndTime: "18:00", PricePerKwh: 0.25},
		{StartTime: "18:00", EndTime: "00:00", PricePerKwh: 0.15},
	}

	err := svc.BulkUpdateSchedules(context.Background(), chargerIDs, schedules)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestBulkUpdateSchedules_EmptyChargerIDs(t *testing.T) {
	repo := newMockRepo()
	svc := NewPricingService(repo)

	schedules := []domain.TOUSchedule{
		{StartTime: "00:00", EndTime: "06:00", PricePerKwh: 0.15},
	}

	err := svc.BulkUpdateSchedules(context.Background(), []string{}, schedules)
	if !errors.Is(err, constants.ErrEmptyChargerID) {
		t.Errorf("expected ErrEmptyChargerID, got %v", err)
	}
}

func TestTimeToMinutes(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"00:00", 0},
		{"06:00", 360},
		{"12:00", 720},
		{"18:30", 1110},
		{"23:59", 1439},
	}

	for _, tt := range tests {
		result := timeToMinutes(tt.input)
		if result != tt.expected {
			t.Errorf("timeToMinutes(%s) = %d, expected %d", tt.input, result, tt.expected)
		}
	}
}
