package api

import (
	"log/slog"
	"net/http"

	"github.com/matheusrb95/fibergraph/internal/response"
)

func logError(r *http.Request, logger *slog.Logger, err error) {
	logger.Error(
		err.Error(),
		"request_method", r.Method,
		"request_url", r.URL.String(),
	)
}

func errorResponse(w http.ResponseWriter, r *http.Request, logger *slog.Logger, status int, message interface{}) {
	err := response.JSON(w, status, response.Envelope{"error": message})
	if err != nil {
		logError(r, logger, err)
		w.WriteHeader(500)
	}
}

func serverErrorResponse(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	logError(r, logger, err)

	message := "The server encountered a problem and could not process your request"
	errorResponse(w, r, logger, http.StatusInternalServerError, message)
}

func badRequestResponse(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	errorResponse(w, r, logger, http.StatusBadRequest, err.Error())
}

func failedValidationResponse(w http.ResponseWriter, r *http.Request, logger *slog.Logger, errors map[string]string) {
	errorResponse(w, r, logger, http.StatusUnprocessableEntity, errors)
}

func notFoundResponse(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	message := "the requested resource could not be found"
	errorResponse(w, r, logger, http.StatusNotFound, message)
}

func conflictResponse(w http.ResponseWriter, r *http.Request, logger *slog.Logger, message string) {
	errorResponse(w, r, logger, http.StatusConflict, message)
}
