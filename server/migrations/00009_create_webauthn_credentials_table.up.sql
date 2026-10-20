-- One row per registered platform-authenticator credential (Face ID, Touch
-- ID, Windows Hello, ...) attached to an existing user (see
-- internal/service/webauthn_credential/). A user may hold several, one per
-- enrolled device; this table never creates a user on its own.
--
-- backup_eligible is the authenticator's "BE" flag at registration time
-- (capable of being backed up/synced, e.g. true for an iCloud Keychain
-- passkey). Per the WebAuthn spec this never changes after registration, and
-- go-webauthn rejects every later login whose assertion doesn't report the
-- same value the authenticator reported here.

CREATE TABLE webauthn_credentials (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id            uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name               text NOT NULL,
    credential_id      bytea NOT NULL,
    public_key         bytea NOT NULL,
    attestation_type   text NOT NULL,
    aaguid             bytea NOT NULL,
    sign_count         bigint NOT NULL DEFAULT 0,
    transports         text[] NOT NULL DEFAULT '{}',
    backup_eligible    boolean NOT NULL DEFAULT false,
    created_at         timestamptz NOT NULL DEFAULT now(),
    last_used_at       timestamptz,

    CONSTRAINT webauthn_credentials_name_not_blank CHECK (btrim(name) <> ''),
    CONSTRAINT webauthn_credentials_credential_id_uq UNIQUE (credential_id)
);

CREATE INDEX webauthn_credentials_user_id_idx ON webauthn_credentials (user_id);
