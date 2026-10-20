-- One table replaces separate income/expense sheets; `type` discriminates.
-- `id` is a bigint identity, exposed by the application (domain) layer as
-- a plain int64.

CREATE TYPE transaction_source AS ENUM ('manual', 'recurring');

CREATE TABLE transactions (
    id                bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    type              transaction_category_type NOT NULL,
    amount            numeric(12,2) NOT NULL,
    currency_code     char(3) NOT NULL REFERENCES currencies (code),
    category_id       uuid NOT NULL,
    user_id           uuid NOT NULL REFERENCES users (id),
    source            transaction_source NOT NULL,
    name              text NOT NULL,
    description       text NOT NULL DEFAULT '',
    created_at        timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT transactions_amount_positive CHECK (amount > 0),
    CONSTRAINT transactions_name_not_blank CHECK (btrim(name) <> ''),

    -- category_id must belong to a category whose type matches this row's
    -- type — a database constraint, not application code.
    CONSTRAINT transactions_category_type_fk
        FOREIGN KEY (category_id, type) REFERENCES transaction_categories (id, type)
);

-- The lookup every "show me this user's transactions over time" query makes.
CREATE INDEX transactions_user_created_idx ON transactions (user_id, created_at);

-- The lookup every "how much did I spend on category X" query makes.
CREATE INDEX transactions_category_idx ON transactions (category_id);

