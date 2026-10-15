package schema_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"jww/internal/adapter/schema"
)

func TestParseDate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{"valid date", "2026-01-15", time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC), false},
		{"leap day", "2024-02-29", time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC), false},
		{"empty string", "", time.Time{}, true},
		{"malformed", "not-a-date", time.Time{}, true},
		{"wrong separator", "2026/01/15", time.Time{}, true},
		{"missing day", "2026-01", time.Time{}, true},
		{"includes time", "2026-01-15T10:00:00Z", time.Time{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := schema.ParseDate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.True(t, tt.want.Equal(got.Time()), "got %v, want %v", got.Time(), tt.want)
		})
	}
}

func TestDate_StringRoundTrip(t *testing.T) {
	d, err := schema.ParseDate("2026-03-07")
	require.NoError(t, err)
	assert.Equal(t, "2026-03-07", d.String())
}

func TestDate_JSONRoundTrip(t *testing.T) {
	d, err := schema.ParseDate("2026-03-07")
	require.NoError(t, err)

	data, err := json.Marshal(d)
	require.NoError(t, err)
	assert.Equal(t, `"2026-03-07"`, string(data))

	var decoded schema.Date
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, d.String(), decoded.String())
}

func TestDate_UnmarshalJSON_InvalidFails(t *testing.T) {
	var d schema.Date
	err := json.Unmarshal([]byte(`"not-a-date"`), &d)
	assert.Error(t, err)
}
