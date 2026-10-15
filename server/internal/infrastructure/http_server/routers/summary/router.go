package summary

import (
	"github.com/go-chi/chi/v5"

	"jww/internal/infrastructure/http_server/kernel"
)

func Mount(_, protected chi.Router, deps *kernel.Deps) {
	protected.Get("/v1/summary", getSummaryHandler(deps))
}
