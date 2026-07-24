package httpapi

import (
	"errors"
	"log"
	"net/http"

	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/service"
)

// errBodyRequired is returned by decodeJSON when the request has no body.
var errBodyRequired = errors.New("request body is required")

// handleServiceError maps a service/db-layer error to the appropriate HTTP
// status code and JSON error body. It's the single place that translates
// sentinel errors into responses, keeping handlers short.
func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, db.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case errors.Is(err, service.ErrSnapshotLocked):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrSnapshotDateExists):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrSnapshotDateInPast):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrPeriodMonthExists):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrNoRateEntry):
		writeError(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, service.ErrNoActivePeriod):
		writeError(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, service.ErrNoSourceMapping):
		writeError(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, service.ErrEnvelopeNotFound):
		writeError(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, service.ErrInvalidCategory):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		log.Printf("internal error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
