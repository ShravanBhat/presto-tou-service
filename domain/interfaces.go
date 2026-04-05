package domain

import (
	"context"
	"time"
)

// Repository defines the data-layer contract.
type Repository interface {
	GetChargerByID(ctx context.Context, chargerID string) (*Charger, error)
	GetPriceAtTime(ctx context.Context, chargerID string, localTime time.Time) (*PricingPeriod, float64, error)
	ReplaceSchedules(ctx context.Context, chargerID string, schedules []TOUSchedule) error
	UpdateSingleSchedule(ctx context.Context, chargerID string, schedule TOUSchedule) error
	GetSchedulesByChargerID(ctx context.Context, chargerID string) ([]TOUSchedule, error)
	BulkReplaceSchedules(ctx context.Context, chargerIDs []string, schedules []TOUSchedule) error
}

// PricingService defines the business-logic contract.
type PricingService interface {
	GetPriceForTime(ctx context.Context, chargerID string, timestamp time.Time) (*PriceResponse, error)
	UpdateSchedules(ctx context.Context, chargerID string, schedules []TOUSchedule) error
	UpdatePartialSchedule(ctx context.Context, chargerID string, schedule TOUSchedule) error
	BulkUpdateSchedules(ctx context.Context, chargerIDs []string, schedules []TOUSchedule) error
	GetSchedules(ctx context.Context, chargerID string) ([]TOUSchedule, error)
}
