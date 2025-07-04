package api

import (
	"log/slog"
	"net/http"

	"github.com/matheusrb95/fibergraph/internal/data"
)

func NewServer(logger *slog.Logger, models *data.Models) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/example", HandleExample(logger, models))

	return mux
}
