package api

import (
	"log/slog"
	"net/http"

	"github.com/matheusrb95/fibergraph/internal/aws"
	"github.com/matheusrb95/fibergraph/internal/data"
)

func NewServer(logger *slog.Logger, models *data.Models, services *aws.Services) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/example", HandleExample(logger, models, services))

	mux.Handle("POST /correlation/{tenant_id}/{project_id}", HandleCorrelation(logger, models, services))

	return mux
}
