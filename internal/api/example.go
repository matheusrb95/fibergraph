package api

import (
	"log/slog"
	"net/http"

	"github.com/matheusrb95/fibergraph/internal/data"
	"github.com/matheusrb95/fibergraph/internal/request"
	"github.com/matheusrb95/fibergraph/internal/response"
)

type Input struct {
	Message string `json:"message"`
}

func HandleExample(logger *slog.Logger, models *data.Models) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input Input
		err := request.DecodeJSON(w, r, &input)
		if err != nil {
			badRequestResponse(w, r, logger, err)
			return
		}

		err = response.JSON(w, http.StatusNoContent, nil)
		if err != nil {
			serverErrorResponse(w, r, logger, err)
		}
	})
}
