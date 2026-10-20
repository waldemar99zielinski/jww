package migrator

import (
	"crypto/sha256"
	"encoding/hex"
)

// Migration is one versioned schema change: an Up script to apply and a
// Down script that reverses it exactly.
type Migration struct {
	Version  string // zero-padded, e.g. "00001"
	Name     string // e.g. "create_users_table"
	Up       string
	Down     string
	Checksum string // sha256 of Up+Down, hex-encoded
}

// Record is one applied migration as stored in schema_migrations.
type Record struct {
	Version  string
	Name     string
	Checksum string
}

func checksum(up, down string) string {
	sum := sha256.Sum256([]byte(up + "\x00" + down))
	return hex.EncodeToString(sum[:])
}
