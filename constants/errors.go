package constants

import "errors"

// Typed errors for the domain layer.
var (
	ErrChargerNotFound       = errors.New("charger not found")
	ErrPriceNotFound         = errors.New("price not found for the given time")
	ErrInvalidTimezone       = errors.New("invalid charger timezone")
	ErrInvalidSchedule       = errors.New("invalid schedule: time periods must cover all 24 hours without gaps or overlaps")
	ErrInvalidPrice          = errors.New("price_per_kwh must be non-negative")
	ErrInvalidTimeFormat     = errors.New("invalid time format, use HH:MM")
	ErrEmptyChargerID        = errors.New("charger_id cannot be empty")
	ErrEmptySchedules        = errors.New("schedules list cannot be empty")
	ErrScheduleNotFound      = errors.New("schedule not found for the given time block")
	ErrScheduleOverlap       = errors.New("schedules must not overlap")
	ErrIncompleteDayCoverage = errors.New("schedules must cover all 24 hours of the day")
	ErrInvalidChargerID      = errors.New("Invalid charger id")
)
