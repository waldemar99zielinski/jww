-- +goose Up
-- Identified by its ISO 4217 code rather than a surrogate id — there's no
-- meaningful existence for a currency apart from that code.

CREATE TABLE currencies (
    code      char(3) PRIMARY KEY,
    name      text NOT NULL,
    is_base   boolean NOT NULL DEFAULT false,

    CONSTRAINT currencies_code_uppercase CHECK (code = upper(code)),
    CONSTRAINT currencies_name_not_blank CHECK (btrim(name) <> '')
);

-- At most one base currency at a time. "At least one" is an application
-- (services-layer) invariant, not something a CHECK/index can express.
CREATE UNIQUE INDEX currencies_single_base_uq ON currencies (is_base) WHERE is_base;

-- +goose Down
DROP INDEX IF EXISTS currencies_single_base_uq;
DROP TABLE IF EXISTS currencies;
