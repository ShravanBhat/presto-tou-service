package domain

// Charger represents an individual EV charging station.
type Charger struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Timezone string `json:"timezone"`
}

// TOUSchedule represents a pricing period for a specific time of day.
// Periods must strictly not cross midnight (use 00:00 as end-time to indicate end-of-day).
type TOUSchedule struct {
	StartTime   string  `json:"start_time"` // Format: "HH:MM"
	EndTime     string  `json:"end_time"`   // Format: "HH:MM"
	PricePerKwh float64 `json:"price_per_kwh"`
}

// PricingPeriod represents the matched pricing period in the response.
type PricingPeriod struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

// PriceResponse defines the data format returned to the client.
type PriceResponse struct {
	ChargerID        string        `json:"charger_id"`
	RequestedTime    string        `json:"requested_time"`
	LocalChargerTime string        `json:"local_charger_time"`
	PricingPeriod    PricingPeriod `json:"pricing_period"`
	PricePerKwh      float64       `json:"price_per_kwh"`
}

// BulkUpdateRequest defines the request body for bulk updating schedules across multiple chargers.
type BulkUpdateRequest struct {
	ChargerIDs []string      `json:"charger_ids"`
	Schedules  []TOUSchedule `json:"schedules"`
}

// ErrorResponse defines the standard error response format.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse defines the standard success response format for mutation operations.
type SuccessResponse struct {
	Message string `json:"message"`
}

// SchedulesResponse wraps the schedules list to avoid returning null when empty.
type SchedulesResponse struct {
	Schedules []TOUSchedule `json:"schedules"`
	Count     int           `json:"count"`
}
