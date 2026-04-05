package service

import (
	"context"
	"fmt"
	"presto_tou_service/constants"
	"presto_tou_service/domain"
	"presto_tou_service/utils"
	"sort"
	"strings"
	"time"
)

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
	localTimeStr := localTime.Format("15:04:05")

	period, price, err := s.repo.GetPriceAtTime(ctx, chargerID, localTimeStr)
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

	if err := utils.ValidateSchedules(schedules); err != nil {
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

	if err := utils.ValidateSingleSchedule(schedule); err != nil {
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

	if err := utils.ValidateSchedules(schedules); err != nil {
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
	// Sort charger IDs in Go before acquiring locks.
	// This guarantees a consistent lock-acquisition order regardless of how
	// PostgreSQL traverses rows, preventing deadlocks when two concurrent
	// bulk-updates target overlapping sets of chargers.
	sortedIDs := make([]string, len(chargerIDs))
	copy(sortedIDs, chargerIDs)
	sort.Strings(sortedIDs)

	return s.repo.BulkReplaceSchedules(ctx, sortedIDs, schedules)
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
