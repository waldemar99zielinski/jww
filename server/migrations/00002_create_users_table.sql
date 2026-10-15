-- +goose Up
-- One row per household member who can log things. Authenticates with a
-- username + password (see internal/service/auth/); display_name is separate,
-- purely for UI.

CREATE TABLE users (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    username           text NOT NULL,
    password_hash      text NOT NULL,
    display_name       text NOT NULL,
    created_at         timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT users_username_not_blank CHECK (btrim(username) <> ''),
    CONSTRAINT users_display_name_not_blank CHECK (btrim(display_name) <> ''),
    CONSTRAINT users_username_uq UNIQUE (username)
);

-- +goose Down
DROP TABLE IF EXISTS users;
