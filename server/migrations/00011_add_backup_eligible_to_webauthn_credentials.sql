-- +goose Up
-- The authenticator's "BE" flag at registration time (capable of being
-- backed up/synced, e.g. true for an iCloud Keychain passkey). Per the
-- WebAuthn spec this never changes after registration, and go-webauthn
-- rejects every later login whose assertion doesn't report the same value
-- the authenticator reported here — this was missing entirely in
-- migration 00010 (always read back as the zero value, false), which broke
-- login for every synced-passkey authenticator (including Face ID/Touch ID
-- with iCloud Keychain) that reports BE=true.
ALTER TABLE webauthn_credentials ADD COLUMN backup_eligible boolean NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE webauthn_credentials DROP COLUMN backup_eligible;
