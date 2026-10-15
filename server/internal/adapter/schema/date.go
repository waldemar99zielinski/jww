package schema

import (
	"fmt"
	"strings"
	"time"
)

// Date is a date-only value (no time-of-day, no zone) serialized as
// "YYYY-MM-DD" — the format the period_start/period_end fields and the
// ?as_of= query param send and receive over JSON/query strings.
type Date time.Time

const dateLayout = "2006-01-02"

func ParseDate(s string) (Date, error) {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return Date{}, fmt.Errorf("invalid date %q, expected YYYY-MM-DD", s)
	}
	return Date(t), nil
}

func (d Date) Time() time.Time { return time.Time(d) }

func (d Date) String() string { return time.Time(d).Format(dateLayout) }

func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.String() + `"`), nil
}

func (d *Date) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	parsed, err := ParseDate(s)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}
