package service

import (
	"context"
	"fmt"
	"presto_tou_service/constants"
	"presto_tou_service/domain"
	"regexp"
	"sort"
	"strings"
	"time"
)

var timeRegex = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)

type pricingService struct {
	repo domain.Repository
}

func NewPricingService(repo domain.Repository) domain.PricingService {
	return &pricingService{repo: repo}
}

func (s *pricingService) GetPriceForTime(ctx context.Context, chargerID string, utcTimestamp time.Time) (*domain.PriceResponse, error) {
	if strings.TrimSpace(chargerID) == constants.Empty {
		return nil, constants.ErrEmptyChargerID
	}

	charger, err := s.repo.GetChargerByID(ctx, chargerID)
	if err != nil {
		return nil, err
	}

	loc, err := time.LoadLocation(charger.Timezone)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", constants.ErrInvalidTimezone, charger.Timezone)
	}

	localTime := utcTimestamp.In(loc)

	period, price, err := s.repo.GetPriceAtTime(ctx, chargerID, localTime)
	if err != nil {
		return nil, err
	}

	return &domain.PriceResponse{
		ChargerID:        chargerID,
		RequestedTime:    utcTimestamp.Format(time.RFC3339),
		LocalChargerTime: localTime.Format(time.RFC3339),
		PricingPeriod: domain.PricingPeriod{
			StartTime: period.StartTime,
			EndTime:   period.EndTime,
		},
		PricePerKwh: price,
	}, nil
}

func (s *pricingService) UpdateSchedules(ctx context.Context, chargerID string, schedules []domain.TOUSchedule) error {
	if strings.TrimSpace(chargerID) == constants.Empty {
		return constants.ErrEmptyChargerID
	}

	if len(schedules) == 0 {
		return constants.ErrEmptySchedules
	}

	if err := validateSchedules(schedules); err != nil {
		return err
	}

	// Verify charger exists before updating
	_, err := s.repo.GetChargerByID(ctx, chargerID)
	if err != nil {
		return err
	}

	return s.repo.ReplaceSchedules(ctx, chargerID, schedules)
}

func (s *pricingService) UpdatePartialSchedule(ctx context.Context, chargerID string, schedule domain.TOUSchedule) error {
	if strings.TrimSpace(chargerID) == constants.Empty {
		return constants.ErrEmptyChargerID
	}

	if err := validateSingleSchedule(schedule); err != nil {
		return err
	}

	// Verify charger exists before updating
	_, err := s.repo.GetChargerByID(ctx, chargerID)
	if err != nil {
		return err
	}

	return s.repo.UpdateSingleSchedule(ctx, chargerID, schedule)
}

func (s *pricingService) BulkUpdateSchedules(ctx context.Context, chargerIDs []string, schedules []domain.TOUSchedule) error {
	if len(chargerIDs) == 0 {
		return constants.ErrEmptyChargerID
	}

	if len(schedules) == 0 {
		return constants.ErrEmptySchedules
	}

	if err := validateSchedules(schedules); err != nil {
		return err
	}

	// Verify all chargers exist before updating
	for _, chargerID := range chargerIDs {
		if strings.TrimSpace(chargerID) == constants.Empty {
			return constants.ErrEmptyChargerID
		}
		_, err := s.repo.GetChargerByID(ctx, chargerID)
		if err != nil {
			return fmt.Errorf("charger %s: %w", chargerID, err)
		}
	}

	return s.repo.BulkReplaceSchedules(ctx, chargerIDs, schedules)
}

func (s *pricingService) GetSchedules(ctx context.Context, chargerID string) ([]domain.TOUSchedule, error) {
	if strings.TrimSpace(chargerID) == constants.Empty {
		return nil, constants.ErrEmptyChargerID
	}

	// Verify charger exists
	_, err := s.repo.GetChargerByID(ctx, chargerID)
	if err != nil {
		return nil, err
	}

	return s.repo.GetSchedulesByChargerID(ctx, chargerID)
}

// validateSchedules checks that schedules are valid:
// - All times are in HH:MM format
// - Prices are non-negative
// - Schedules cover all 24 hours without gaps or overlaps
func validateSchedules(schedules []domain.TOUSchedule) error {
	if len(schedules) == 0 {
		return constants.ErrEmptySchedules
	}

	// Validate individual schedules first
	for _, s := range schedules {
		if err := validateSingleSchedule(s); err != nil {
			return err
		}
	}

	// Check for 24-hour coverage and no overlaps
	// We normalize all schedules to a single day timeline and check coverage
	type timeSegment struct {
		startMinutes int
		endMinutes   int
	}

	var segments []timeSegment
	const minutesInDay = 24 * 60

	for _, s := range schedules {
		startMin := timeToMinutes(s.StartTime)
		endMin := timeToMinutes(s.EndTime)

		if s.EndTime == "00:00" {
			endMin = minutesInDay
		}

		segments = append(segments, timeSegment{startMin, endMin})
	}

	// Sort segments by start time
	sort.Slice(segments, func(i, j int) bool {
		return segments[i].startMinutes < segments[j].startMinutes
	})

	// Check for gaps and overlaps
	for i, seg := range segments {
		if seg.startMinutes >= seg.endMinutes {
			return fmt.Errorf("%w: segment %d has invalid range", constants.ErrInvalidSchedule, i)
		}

		// Check for overlap with next segment
		if i < len(segments)-1 {
			if seg.endMinutes > segments[i+1].startMinutes {
				return fmt.Errorf("%w: segment %d overlaps with segment %d", constants.ErrScheduleOverlap, i, i+1)
			}
			// Check for gap
			if seg.endMinutes < segments[i+1].startMinutes {
				return fmt.Errorf("%w: gap between %d and %d minutes", constants.ErrIncompleteDayCoverage, seg.endMinutes, segments[i+1].startMinutes)
			}
		}
	}

	// Check that the first segment starts at 0 and the last ends at 1440 (24*60)
	if segments[0].startMinutes != 0 {
		return fmt.Errorf("%w: first segment starts at %d minutes, expected 0", constants.ErrIncompleteDayCoverage, segments[0].startMinutes)
	}
	if segments[len(segments)-1].endMinutes != minutesInDay {
		return fmt.Errorf("%w: last segment ends at %d minutes, expected %d", constants.ErrIncompleteDayCoverage, segments[len(segments)-1].endMinutes, minutesInDay)
	}

	return nil
}

// validateSingleSchedule validates a single schedule entry.
func validateSingleSchedule(s domain.TOUSchedule) error {
	if !timeRegex.MatchString(s.StartTime) {
		return fmt.Errorf("%w: start_time '%s'", constants.ErrInvalidTimeFormat, s.StartTime)
	}
	if !timeRegex.MatchString(s.EndTime) {
		return fmt.Errorf("%w: end_time '%s'", constants.ErrInvalidTimeFormat, s.EndTime)
	}
	if s.PricePerKwh < 0 {
		return constants.ErrInvalidPrice
	}
	if s.StartTime == s.EndTime {
		return fmt.Errorf("%w: start_time and end_time cannot be the same", constants.ErrInvalidSchedule)
	}

	startMin := timeToMinutes(s.StartTime)
	endMin := timeToMinutes(s.EndTime)
	if s.EndTime == "00:00" {
		endMin = 24 * 60
	}

	if startMin >= endMin {
		return fmt.Errorf("%w: start_time must be before end_time", constants.ErrInvalidSchedule)
	}

	return nil
}

// timeToMinutes converts a HH:MM time string to minutes since midnight.
func timeToMinutes(t string) int {
	parts := strings.Split(t, ":")
	if len(parts) != 2 {
		return 0
	}
	var hours, minutes int
	fmt.Sscanf(parts[0], "%d", &hours)
	fmt.Sscanf(parts[1], "%d", &minutes)
	return hours*60 + minutes
}
