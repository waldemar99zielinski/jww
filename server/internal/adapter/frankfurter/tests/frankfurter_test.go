package frankfurter_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/frankfurter"
	"jww/internal/domain"
)

func appContext() domain.AppContext {
	var appCtx domain.AppContext
	appCtx.SetContext(context.Background())
	return appCtx
}

func newClient(serverURL string) *frankfurter.Client {
	return frankfurter.NewClient(frankfurter.Config{BaseURL: serverURL, TimeoutSeconds: 5})
}

func TestClient_Check(t *testing.T) {
	var path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	err := newClient(srv.URL).Check(appContext())

	require.NoError(t, err)
	assert.Equal(t, "/v2/currencies", path)
}

// rateRow mirrors one element of the /v2/rates array so tests can build
// responses without hand-writing JSON. Rate is written verbatim as a bare JSON
// number (never quoted), matching the API.
type rateRow struct {
	Date  string
	Base  string
	Quote string
	Rate  string
}

func writeRows(t *testing.T, w http.ResponseWriter, rows []rateRow) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	_, err := fmt.Fprint(w, "[")
	require.NoError(t, err)
	for i, row := range rows {
		if i > 0 {
			_, err = fmt.Fprint(w, ",")
			require.NoError(t, err)
		}
		_, err = fmt.Fprintf(w, `{"date":%q,"base":%q,"quote":%q,"rate":%s}`, row.Date, row.Base, row.Quote, row.Rate)
		require.NoError(t, err)
	}
	_, err = fmt.Fprint(w, "]")
	require.NoError(t, err)
}

func TestClient_Rates_OneRequestPerCurrencyWithTargetDate(t *testing.T) {
	type request struct{ base, quotes, date string }
	var requests []request

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		base := query.Get("base")
		requests = append(requests, request{base: base, quotes: query.Get("quotes"), date: query.Get("date")})

		rows := make([]rateRow, 0)
		for _, quote := range strings.Split(query.Get("quotes"), ",") {
			if quote == "" {
				continue
			}
			rows = append(rows, rateRow{Date: query.Get("date"), Base: base, Quote: quote, Rate: "1.0834"})
		}
		writeRows(t, w, rows)
	}))
	defer srv.Close()

	asOf := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	fetch, err := newClient(srv.URL).Rates(appContext(), []string{"EUR", "USD"}, asOf)

	require.NoError(t, err)
	require.Len(t, requests, 2, "one request per currency as base")
	assert.Equal(t, "2026-10-14", requests[0].date)
	assert.Equal(t, "2026-10-14", requests[1].date)
	assert.ElementsMatch(t, []string{"EUR", "USD"}, []string{requests[0].base, requests[1].base})

	require.Len(t, fetch.Rates, 2, "the full directed matrix among two currencies")
	assert.Empty(t, fetch.Skipped)
	for _, rate := range fetch.Rates {
		assert.Equal(t, domain.ExchangeRate(10834), rate.Rate)
		assert.Equal(t, asOf, rate.RecordedAt, "the echoed business date becomes RecordedAt at midnight UTC")
	}
}

func TestClient_Rates_ZeroAsOfOmitsDate(t *testing.T) {
	var sawDate string
	var sawDateParam bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawDate = r.URL.Query().Get("date")
		_, sawDateParam = r.URL.Query()["date"]
		writeRows(t, w, rowsForRequest(r, "2026-10-15", "1.1199"))
	}))
	defer srv.Close()

	fetch, err := newClient(srv.URL).Rates(appContext(), []string{"EUR", "USD"}, time.Time{})

	require.NoError(t, err)
	assert.False(t, sawDateParam, "a zero asOf must not send a date parameter")
	assert.Empty(t, sawDate)
	require.Len(t, fetch.Rates, 2)
}

func TestClient_Rates_OmittedQuoteBecomesASkip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		base := query.Get("base")
		if base == "USD" {
			// Provider has no rate for USD→PLN on this day: empty array.
			writeRows(t, w, nil)
			return
		}
		writeRows(t, w, []rateRow{{Date: query.Get("date"), Base: base, Quote: "USD", Rate: "0.2562"}})
	}))
	defer srv.Close()

	asOf := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	fetch, err := newClient(srv.URL).Rates(appContext(), []string{"USD", "PLN"}, asOf)

	require.NoError(t, err)
	require.Len(t, fetch.Rates, 1)
	assert.Equal(t, "PLN", fetch.Rates[0].FromCurrencyCode)
	require.Len(t, fetch.Skipped, 1, "every requested quote the provider omits is reported, never dropped")
	assert.Equal(t, "USD", fetch.Skipped[0].From)
	assert.Equal(t, "PLN", fetch.Skipped[0].To)
}

func TestClient_Rates_EmptyArrayBecomesASkipPerQuote(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeRows(t, w, nil)
	}))
	defer srv.Close()

	fetch, err := newClient(srv.URL).Rates(appContext(), []string{"EUR", "USD"}, time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC))

	require.NoError(t, err)
	assert.Empty(t, fetch.Rates)
	assert.Len(t, fetch.Skipped, 2)
}

func TestClient_Rates_Errors(t *testing.T) {
	t.Run("500 aborts the run", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		_, err := newClient(srv.URL).Rates(appContext(), []string{"EUR", "USD"}, time.Time{})
		assert.Error(t, err)
	})

	t.Run("malformed JSON aborts the run", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, "{not json")
		}))
		defer srv.Close()

		_, err := newClient(srv.URL).Rates(appContext(), []string{"EUR", "USD"}, time.Time{})
		assert.Error(t, err)
	})
}

func TestClient_Rates_ParsesFiveDecimalRate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeRows(t, w, rowsForRequest(r, "2026-10-14", "1.08342"))
	}))
	defer srv.Close()

	fetch, err := newClient(srv.URL).Rates(appContext(), []string{"EUR", "USD"}, time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC))

	require.NoError(t, err)
	require.Len(t, fetch.Rates, 2)
	for _, rate := range fetch.Rates {
		assert.Equal(t, domain.ExchangeRate(10834), rate.Rate, "the decimal is parsed, never decoded as float64")
	}
}

// rowsForRequest builds one response row per requested quote, echoing the
// requested base and date — enough for tests that don't care about the exact
// pair.
func rowsForRequest(r *http.Request, date, rate string) []rateRow {
	query := r.URL.Query()
	base := query.Get("base")
	if date == "" {
		date = query.Get("date")
	}
	rows := make([]rateRow, 0)
	for _, quote := range strings.Split(query.Get("quotes"), ",") {
		if quote == "" {
			continue
		}
		rows = append(rows, rateRow{Date: date, Base: base, Quote: quote, Rate: rate})
	}
	return rows
}
