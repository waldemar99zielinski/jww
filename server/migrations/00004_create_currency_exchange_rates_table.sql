-- +goose Up
-- Append-only log: "1 unit of from_currency_code was worth `rate` units of
-- to_currency_code as of recorded_at." Never updated in place — "the rate
-- right now" is just the latest row per pair, and "what was the rate on a
-- given date" is a plain query (latest recorded_at <= that date) instead of
-- a manually overwritten single cell.
--
-- Identified by (from_currency_code, to_currency_code, recorded_at), not a
-- surrogate id — multiple historical rates exist per pair, so the pair alone
-- isn't unique, but the pair plus the timestamp it was recorded at is.

CREATE TABLE currency_exchange_rates (
    from_currency_code   char(3) NOT NULL REFERENCES currencies (code),
    to_currency_code      char(3) NOT NULL REFERENCES currencies (code),
    rate                   numeric(18,4) NOT NULL,
    recorded_at            timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT currency_exchange_rates_pk PRIMARY KEY (from_currency_code, to_currency_code, recorded_at),
    CONSTRAINT currency_exchange_rates_distinct_currencies CHECK (from_currency_code <> to_currency_code),
    CONSTRAINT currency_exchange_rates_rate_positive CHECK (rate > 0)
);

-- +goose Down
DROP TABLE IF EXISTS currency_exchange_rates;
