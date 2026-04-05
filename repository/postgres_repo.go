package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"presto_tou_service/constants"
	"presto_tou_service/domain"

	"github.com/lib/pq"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return constants.ErrChargerNotFound
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "22P02" {
		return constants.ErrInvalidChargerID
	}
	return err
}

func (r *PostgresRepo) GetChargerByID(ctx context.Context, chargerID string) (*domain.Charger, error) {
	var c domain.Charger
	query := `SELECT id, name, timezone FROM chargers WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, chargerID).Scan(&c.ID, &c.Name, &c.Timezone)
	if err != nil {
		mappedErr := r.mapError(err)
		if errors.Is(mappedErr, constants.ErrChargerNotFound) || errors.Is(mappedErr, constants.ErrInvalidChargerID) {
			return nil, mappedErr
		}
		return nil, fmt.Errorf("failed to get charger: %w", err)
	}
	return &c, nil
}

func (r *PostgresRepo) GetPriceAtTime(ctx context.Context, chargerID string, timeStr string) (*domain.PricingPeriod, float64, error) {
	var period domain.PricingPeriod
	var price float64

	// Query for non-overnight schedules first, then overnight schedules.
	// For non-overnight: start_time <= current_time < end_time
	// For overnight: start_time > end_time, and current_time >= start_time OR current_time < end_time
	query := `
		SELECT start_time::text, end_time::text, price_per_kwh
		FROM tou_schedules
		WHERE charger_id = $1
		  AND (
			start_time <= $2 AND 
			(end_time > $2 OR end_time = '00:00:00')
		  )
		ORDER BY start_time
		LIMIT 1`

	err := r.db.QueryRowContext(ctx, query, chargerID, timeStr).Scan(&period.StartTime, &period.EndTime, &price)
	if err != nil {
		mappedErr := r.mapError(err)
		if errors.Is(mappedErr, constants.ErrChargerNotFound) || errors.Is(mappedErr, constants.ErrPriceNotFound) || errors.Is(mappedErr, constants.ErrInvalidChargerID) {
			// Special handling: sql.ErrNoRows for price at time should actually be ErrPriceNotFound
			if errors.Is(err, sql.ErrNoRows) {
				return nil, 0, constants.ErrPriceNotFound
			}
			return nil, 0, mappedErr
		}
		return nil, 0, fmt.Errorf("failed to get price: %w", err)
	}

	return &period, price, nil
}

func (r *PostgresRepo) ReplaceSchedules(ctx context.Context, chargerID string, schedules []domain.TOUSchedule) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Lock charger to avoid race conditions
	_, err = tx.ExecContext(ctx, `SELECT id FROM chargers WHERE id = $1 FOR UPDATE`, chargerID)
	if err != nil {
		mappedErr := r.mapError(err)
		if errors.Is(mappedErr, constants.ErrInvalidChargerID) || errors.Is(mappedErr, constants.ErrChargerNotFound) {
			return mappedErr
		}
		return fmt.Errorf("failed to lock charger: %w", mappedErr)
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM tou_schedules WHERE charger_id = $1`, chargerID)
	if err != nil {
		return fmt.Errorf("failed to delete schedules: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO tou_schedules (charger_id, start_time, end_time, price_per_kwh) VALUES ($1, $2, $3, $4)`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, s := range schedules {
		_, err = stmt.ExecContext(ctx, chargerID, s.StartTime, s.EndTime, s.PricePerKwh)
		if err != nil {
			return fmt.Errorf("failed to insert schedule: %w", err)
		}
	}

	return tx.Commit()
}

func (r *PostgresRepo) UpdateSingleSchedule(ctx context.Context, chargerID string, schedule domain.TOUSchedule) error {
	query := `
		UPDATE tou_schedules 
		SET price_per_kwh = $1
		WHERE charger_id = $2 
		  AND start_time = $3 
		  AND end_time = $4
	`

	result, err := r.db.ExecContext(ctx, query, schedule.PricePerKwh, chargerID, schedule.StartTime, schedule.EndTime)
	if err != nil {
		mappedErr := r.mapError(err)
		if errors.Is(mappedErr, constants.ErrInvalidChargerID) {
			return mappedErr
		}
		return fmt.Errorf("failed to update schedule: %w", mappedErr)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return constants.ErrScheduleNotFound
	}

	return nil
}

func (r *PostgresRepo) GetSchedulesByChargerID(ctx context.Context, chargerID string) ([]domain.TOUSchedule, error) {
	query := `
		SELECT start_time::text, end_time::text, price_per_kwh
		FROM tou_schedules
		WHERE charger_id = $1
		ORDER BY start_time
	`

	rows, err := r.db.QueryContext(ctx, query, chargerID)
	if err != nil {
		mappedErr := r.mapError(err)
		if errors.Is(mappedErr, constants.ErrInvalidChargerID) {
			return nil, mappedErr
		}
		return nil, fmt.Errorf("failed to query schedules: %w", mappedErr)
	}
	defer rows.Close()

	var schedules []domain.TOUSchedule
	for rows.Next() {
		var s domain.TOUSchedule
		if err := rows.Scan(&s.StartTime, &s.EndTime, &s.PricePerKwh); err != nil {
			return nil, fmt.Errorf("failed to scan schedule: %w", err)
		}
		schedules = append(schedules, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schedules: %w", err)
	}

	return schedules, nil
}

func (r *PostgresRepo) BulkReplaceSchedules(ctx context.Context, sortedIDs []string, schedules []domain.TOUSchedule) error {

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Lock chargers in sorted order to prevent deadlocks.
	_, err = tx.ExecContext(ctx, `SELECT id FROM chargers WHERE id = ANY($1) ORDER BY id FOR UPDATE`, pq.Array(sortedIDs))
	if err != nil {
		return fmt.Errorf("failed to lock chargers: %w", err)
	}

	// Delete existing schedules for all specified chargers
	_, err = tx.ExecContext(ctx, `DELETE FROM tou_schedules WHERE charger_id = ANY($1)`, pq.Array(sortedIDs))
	if err != nil {
		return fmt.Errorf("failed to delete schedules: %w", err)
	}

	// Build bulk insert query
	query := "INSERT INTO tou_schedules (charger_id, start_time, end_time, price_per_kwh) VALUES "
	var vals []interface{}
	i := 1
	for _, chargerID := range sortedIDs {
		for _, s := range schedules {
			query += fmt.Sprintf("($%d, $%d, $%d, $%d),", i, i+1, i+2, i+3)
			vals = append(vals, chargerID, s.StartTime, s.EndTime, s.PricePerKwh)
			i += 4
		}
	}
	query = query[:len(query)-1] // remove trailing comma

	_, err = tx.ExecContext(ctx, query, vals...)
	if err != nil {
		return fmt.Errorf("failed to bulk insert schedules: %w", err)
	}

	return tx.Commit()
}
