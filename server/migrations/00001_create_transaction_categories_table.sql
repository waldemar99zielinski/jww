-- +goose Up
-- Categories for both income and expense transactions, discriminated by `type`.
-- A category name may repeat across types (e.g. "Gifts" is legitimately both an
-- income category and an expense category) but not within one.
--
-- (id, type) is additionally unique so later migrations (transactions,
-- recurring_templates) can enforce "this row's type must match its category's
-- type" with a composite foreign key — a database constraint, not application
-- code.

CREATE EXTENSION IF NOT EXISTS pgcrypto;  -- gen_random_uuid(); harmless no-op on PG13+ where it's built in

CREATE TYPE transaction_category_type AS ENUM ('income', 'expense');

CREATE TABLE transaction_categories (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    type        transaction_category_type NOT NULL,
    name        text NOT NULL,
    active      boolean NOT NULL DEFAULT true,
    sort_order  smallint NOT NULL DEFAULT 0,

    CONSTRAINT transaction_categories_name_not_blank CHECK (btrim(name) <> ''),
    CONSTRAINT transaction_categories_sort_order_non_negative CHECK (sort_order >= 0),

    -- Enables composite FKs from other tables; see migration note above.
    CONSTRAINT transaction_categories_id_type_uq UNIQUE (id, type)
);

-- Same name allowed once per type, not globally ("Gifts" can be both).
CREATE UNIQUE INDEX transaction_categories_type_name_uq ON transaction_categories (type, name);

-- The lookup every "show me the category picker" call will make.
CREATE INDEX transaction_categories_type_active_idx ON transaction_categories (type, active);

-- +goose Down
DROP INDEX IF EXISTS transaction_categories_type_active_idx;
DROP INDEX IF EXISTS transaction_categories_type_name_uq;
DROP TABLE IF EXISTS transaction_categories;
DROP TYPE IF EXISTS transaction_category_type;
