package domain

import (
	"math"
	"testing"
)

func TestParseMoneyAndString(t *testing.T) {
	cases := map[string]string{
		"12.34": "12.34",
		"-7":    "-7.00",
		"0":     "0.00",
		"100.5": "100.50",
		"-0.5":  "-0.50",
		"12.3":  "12.30",
		"12":    "12.00",
		"+5.25": "5.25",
	}
	for in, want := range cases {
		m, err := ParseMoney(in)
		if err != nil {
			t.Fatalf("ParseMoney(%q): %v", in, err)
		}
		if got := m.String(); got != want {
			t.Errorf("ParseMoney(%q).String() = %q, want %q", in, got, want)
		}
	}
}

func TestParseMoneyRoundsHalfAwayFromZero(t *testing.T) {
	cases := map[string]string{
		"12.345":  "12.35",
		"12.344":  "12.34",
		"-12.345": "-12.35",
		"0.005":   "0.01",
		"9.995":   "10.00",
	}
	for in, want := range cases {
		m, err := ParseMoney(in)
		if err != nil {
			t.Fatalf("ParseMoney(%q): %v", in, err)
		}
		if got := m.String(); got != want {
			t.Errorf("ParseMoney(%q).String() = %q, want %q", in, got, want)
		}
	}
}

func TestParseMoneyInvalid(t *testing.T) {
	for _, s := range []string{"", "abc", "1.2.3", "1-2"} {
		if _, err := ParseMoney(s); err == nil {
			t.Errorf("ParseMoney(%q) expected error, got nil", s)
		}
	}
}

func TestMoneyArithmeticIsPlainInt(t *testing.T) {
	a := Money(1050) // 10.50
	b := Money(250)  // 2.50
	if got := a + b; got.String() != "13.00" {
		t.Errorf("10.50 + 2.50 = %s, want 13.00", got)
	}
	if got := a - b; got.String() != "8.00" {
		t.Errorf("10.50 - 2.50 = %s, want 8.00", got)
	}
	if !(b < a) {
		t.Error("2.50 should be less than 10.50")
	}
}

func TestParseExchangeRateAndString(t *testing.T) {
	r, err := ParseExchangeRate("4.1234")
	if err != nil {
		t.Fatalf("ParseExchangeRate: %v", err)
	}
	if got := r.String(); got != "4.1234" {
		t.Errorf("String() = %q, want 4.1234", got)
	}
}

func TestExchangeRateConvert(t *testing.T) {
	amount := Money(1000) // 10.00
	rate, err := ParseExchangeRate("4.1234")
	if err != nil {
		t.Fatalf("ParseExchangeRate: %v", err)
	}
	// 10.00 * 4.1234 = 41.234, rounds to 41.23.
	got := rate.Convert(amount)
	if want := Money(4123); got != want {
		t.Errorf("Convert = %d (%s), want %d (41.23)", got, got, want)
	}
}

func TestExchangeRateConvertRoundsHalfAwayFromZero(t *testing.T) {
	// 1 unit at a rate chosen so the product lands exactly on a half-cent.
	amount := Money(1)
	rate := ExchangeRate(5_000) // 0.5
	got := rate.Convert(amount)
	if want := Money(1); got != want {
		t.Errorf("Convert = %d, want %d", got, want)
	}
}

func TestExchangeRateConvertNegativeSigns(t *testing.T) {
	rate, err := ParseExchangeRate("4.1234")
	if err != nil {
		t.Fatalf("ParseExchangeRate: %v", err)
	}
	// 10.00 * 4.1234 = 41.234, rounds to 41.23 in magnitude regardless of
	// which operand carries the sign.
	if got, want := rate.Convert(Money(-1000)), Money(-4123); got != want {
		t.Errorf("negative amount: Convert = %d, want %d", got, want)
	}
	if got, want := (-rate).Convert(Money(1000)), Money(-4123); got != want {
		t.Errorf("negative rate: Convert = %d, want %d", got, want)
	}
	if got, want := (-rate).Convert(Money(-1000)), Money(4123); got != want {
		t.Errorf("negative amount and rate: Convert = %d, want %d", got, want)
	}
}

func TestExchangeRateConvertZero(t *testing.T) {
	if got := ExchangeRate(0).Convert(Money(100)); got != 0 {
		t.Errorf("Convert with zero rate = %d, want 0", got)
	}
	rate, _ := ParseExchangeRate("1.5")
	if got := rate.Convert(Money(0)); got != 0 {
		t.Errorf("Convert of zero amount = %d, want 0", got)
	}
}

func TestExchangeRateConvertPanicsOnOverflow(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Convert of two int64-max values should panic on overflow")
		}
	}()
	ExchangeRate(math.MaxInt64).Convert(Money(math.MaxInt64))
}

func TestMoneyJSONRoundTripUsesCents(t *testing.T) {
	m := Money(1234)
	data, err := m.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	if string(data) != "1234" {
		t.Errorf("MarshalJSON = %s, want bare cents 1234", data)
	}
	var parsed Money
	if err := parsed.UnmarshalJSON(data); err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}
	if parsed != m {
		t.Errorf("round trip = %d, want %d", parsed, m)
	}
	var quoted Money
	if err := quoted.UnmarshalJSON([]byte(`"1234"`)); err != nil {
		t.Fatalf("UnmarshalJSON quoted: %v", err)
	}
	if quoted != m {
		t.Errorf("quoted round trip = %d, want %d", quoted, m)
	}
}

func TestMoneyJSONRejectsMajorUnitDecimals(t *testing.T) {
	var m Money
	if err := m.UnmarshalJSON([]byte(`41.27`)); err == nil {
		t.Fatal("decimal money must be rejected; the wire contract is integer cents")
	}
}
