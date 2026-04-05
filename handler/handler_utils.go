package handler

import (
	"errors"
	"net/http"
	"presto_tou_service/constants"
)

// httpStatusForError maps domain errors to appropriate HTTP status codes.
func httpStatusForError(err error) int {
	switch {
	case errors.Is(err, constants.ErrChargerNotFound):
		return http.StatusNotFound
	case errors.Is(err, constants.ErrPriceNotFound):
		return http.StatusNotFound
	case errors.Is(err, constants.ErrScheduleNotFound):
		return http.StatusNotFound
	case errors.Is(err, constants.ErrEmptyChargerID):
		return http.StatusBadRequest
	case errors.Is(err, constants.ErrEmptySchedules):
		return http.StatusBadRequest
	case errors.Is(err, constants.ErrInvalidTimeFormat):
		return http.StatusBadRequest
	case errors.Is(err, constants.ErrInvalidPrice):
		return http.StatusBadRequest
	case errors.Is(err, constants.ErrInvalidSchedule):
		return http.StatusBadRequest
	case errors.Is(err, constants.ErrScheduleOverlap):
		return http.StatusBadRequest
	case errors.Is(err, constants.ErrIncompleteDayCoverage):
		return http.StatusBadRequest
	case errors.Is(err, constants.ErrInvalidTimezone):
		return http.StatusInternalServerError
	case errors.Is(err, constants.ErrInvalidChargerID):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
