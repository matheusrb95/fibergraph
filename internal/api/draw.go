package api

import (
	"log/slog"
	"net/http"

	"github.com/matheusrb95/fibergraph/internal/core"
	"github.com/matheusrb95/fibergraph/internal/data"
	"github.com/matheusrb95/fibergraph/internal/response"
)

func HandleDraw(logger *slog.Logger, models *data.Models) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := core.Run(data.Network3())
		if err != nil {
			serverErrorResponse(w, r, logger, err)
			return
		}

		err = response.JSON(w, http.StatusCreated, response.Envelope{"message": "draw done"})
		if err != nil {
			serverErrorResponse(w, r, logger, err)
		}
	})
}
