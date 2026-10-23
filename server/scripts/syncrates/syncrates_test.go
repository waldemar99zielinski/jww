package syncrates

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/domain"
	"jww/internal/service/workflow"
)

// fakeSyncer records the days it was asked for and returns canned results,
// keyed by YYYY-MM-DD, so the range logic can be exercised without Postgres or
// the provider.
type fakeSyncer struct {
	calls   []time.Time
	results map[string]workflow.SyncResult
	errs    map[string]error
	onSync  func()
}

func (f *fakeSyncer) Sync(_ domain.AppContext, asOf time.Time) (workflow.SyncResult, error) {
	f.calls = append(f.calls, asOf)
	if f.onSync != nil {
		f.onSync()
	}
	key := asOf.Format("2006-01-02")
	return f.results[key], f.errs[key]
}

func TestClient_SyncRange_RejectsInvertedRange(t *testing.T) {
	sync := &fakeSyncer{}
	client := newClient(sync)

	_, err := client.SyncRange(newAppContext(context.Background()), date(2026, 10, 19), date(2026, 10, 16), false)

	require.ErrorContains(t, err, "end date 2026-10-16 is before start date 2026-10-19")
	assert.Empty(t, sync.calls, "an invalid range must not call the workflow")
}

func TestClient_SyncRange_RejectsFutureEndDate(t *testing.T) {
	sync := &fakeSyncer{}
	client := newClient(sync)
	tomorrow := dateOnly(time.Now().UTC().AddDate(0, 0, 1))

	_, err := client.SyncRange(newAppContext(context.Background()), tomorrow.AddDate(0, 0, -2), tomorrow, false)

	require.ErrorContains(t, err, "is in the future")
	assert.Empty(t, sync.calls, "a future range must not call the workflow")
}

func TestClient_SyncRange_SkipsWeekendsByDefault(t *testing.T) {
	// 2026-10-16 is a Friday; the range ends on Monday the 19th.
	sync := &fakeSyncer{results: map[string]workflow.SyncResult{
		"2026-10-16": {Currencies: 3, Fetched: 6, Inserted: 6},
		"2026-10-19": {Currencies: 3, Fetched: 6, Inserted: 6},
	}}
	client := newClient(sync)

	result, err := client.SyncRange(newAppContext(context.Background()), date(2026, 10, 16), date(2026, 10, 19), false)

	require.NoError(t, err)
	assert.Len(t, sync.calls, 2, "Saturday and Sunday must not reach the provider")
	assert.Equal(t, 2, result.WeekendDays)
	assert.Equal(t, 2, result.Synced)
	assert.Len(t, result.Days, 4, "skipped days stay in the report")
	for _, day := range result.Days {
		if day.Date.Weekday() == time.Saturday || day.Date.Weekday() == time.Sunday {
			assert.True(t, day.Weekend, "%s should be marked as a weekend", day.Date.Format("2006-01-02"))
		}
	}
}

func TestClient_SyncRange_IncludesWeekendsWhenAsked(t *testing.T) {
	sync := &fakeSyncer{}
	client := newClient(sync)

	result, err := client.SyncRange(newAppContext(context.Background()), date(2026, 10, 16), date(2026, 10, 19), true)

	require.NoError(t, err)
	assert.Len(t, sync.calls, 4)
	assert.Zero(t, result.WeekendDays)
	assert.Equal(t, 4, result.Synced)
}

func TestClient_SyncRange_AggregatesCountsAndContinuesPastFailedDays(t *testing.T) {
	sync := &fakeSyncer{
		results: map[string]workflow.SyncResult{
			"2026-10-16": {
				Currencies: 3, Fetched: 6, Inserted: 4, Updated: 2, Skipped: 1,
				Warnings: []string{"no rate for USD→EUR on 2026-10-16"},
			},
			"2026-10-19": {Currencies: 3, Fetched: 6, Inserted: 6},
		},
		errs: map[string]error{"2026-10-19": errors.New("provider down")},
	}
	client := newClient(sync)

	result, err := client.SyncRange(newAppContext(context.Background()), date(2026, 10, 16), date(2026, 10, 19), false)

	require.ErrorContains(t, err, "1 of 2 day(s) failed")
	assert.Len(t, sync.calls, 2, "a failed day must not abort the rest of the range")
	assert.Equal(t, 1, result.Synced)
	assert.Equal(t, 1, result.Failed)
	assert.Equal(t, 4, result.Inserted, "failed days must not contribute counts")
	assert.Equal(t, 2, result.Updated)
	assert.Equal(t, 1, result.SkippedPairs)
	assert.Equal(t, []string{"no rate for USD→EUR on 2026-10-16"}, result.Days[0].Result.Warnings,
		"per-day warnings must reach the caller")
}

func TestClient_SyncRange_ComparesCalendarDaysInUTC(t *testing.T) {
	sync := &fakeSyncer{}
	client := newClient(sync)
	from := time.Date(2026, 10, 19, 0, 30, 0, 0, time.FixedZone("UTC+14", 14*60*60))

	_, err := client.SyncRange(newAppContext(context.Background()), from, date(2026, 10, 19), false)

	require.NoError(t, err)
	require.Len(t, sync.calls, 1)
	assert.Equal(t, date(2026, 10, 19), sync.calls[0], "the time of day must not shift the calendar day")
}

// date builds a UTC calendar day, the same normalized form SyncRange uses.
func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

// newAppContext mirrors the app context the command builds: SyncRange reads
// cancellation through it, so a zero value (nil context) is not enough.
func newAppContext(ctx context.Context) domain.AppContext {
	var appCtx domain.AppContext
	appCtx.SetContext(ctx)
	return appCtx
}

func TestClient_SyncRange_StopsWhenTheContextIsCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	sync := &fakeSyncer{onSync: cancel}
	client := newClient(sync)

	result, err := client.SyncRange(newAppContext(ctx), date(2026, 10, 19), date(2026, 10, 21), false)

	require.ErrorContains(t, err, "range sync interrupted after 1 day(s)")
	assert.ErrorIs(t, err, context.Canceled)
	assert.Len(t, sync.calls, 1, "cancellation must not walk the remaining days")
	assert.Equal(t, 1, result.Synced)
}
