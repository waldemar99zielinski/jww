package summary

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"jww/internal/adapter/schema"
	"jww/internal/domain"
	"jww/internal/infrastructure/http_server/kernel"
)

var (
	errMissingCurrencyCode = errors.New("currency_code is required")
	errInvalidPeriod       = errors.New("period must be one of 'day', 'week', 'month', 'year'")
)

func getSummaryHandler(deps *kernel.Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := kernel.RequestContext(r)

		q := r.URL.Query()

		periodEnd, err := schema.ParseDate(q.Get("period_end"))
		if err != nil {
			kernel.WriteBadRequest(w, err)
			return
		}
		currencyCode := q.Get("currency_code")
		if currencyCode == "" {
			kernel.WriteBadRequest(w, errMissingCurrencyCode)
			return
		}

		var period *domain.SummaryPeriod
		if v := q.Get("period"); v != "" {
			p := domain.SummaryPeriod(v)
			if !p.Valid() {
				kernel.WriteBadRequest(w, errInvalidPeriod)
				return
			}
			period = &p
		}

		var periodStart *time.Time
		if v := q.Get("period_start"); v != "" {
			d, err := schema.ParseDate(v)
			if err != nil {
				kernel.WriteBadRequest(w, err)
				return
			}
			t := d.Time()
			periodStart = &t
		}

		var userID *uuid.UUID
		if v := q.Get("user_id"); v != "" {
			id, err := uuid.Parse(v)
			if err != nil {
				kernel.WriteBadRequest(w, err)
				return
			}
			userID = &id
		}

		summary, err := deps.SummaryWorkflow.Generate(appCtx, period, periodEnd.Time(), currencyCode, periodStart, userID)
		if err != nil {
			kernel.WriteError(w, err)
			return
		}
		kernel.WriteJSON(w, http.StatusOK, schema.SummaryToResponse(*summary))
	}
}
