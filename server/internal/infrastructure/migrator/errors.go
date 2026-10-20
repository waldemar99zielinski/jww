package migrator

import (
	"errors"
	"fmt"
)

// ErrChecksumMismatch is wrapped with migration-specific detail and returned
// by Up/Status when a previously-applied migration's files no longer match
// the checksum recorded at apply time.
var ErrChecksumMismatch = errors.New("migrator: migration file changed after being applied")

// ErrMissingFile is wrapped with migration-specific detail and returned when
// a previously-applied migration's version has no corresponding files on
// disk anymore.
var ErrMissingFile = errors.New("migrator: a previously applied migration's files are missing")

// verify checks that every already-applied record still has a matching,
// unmodified migration on disk. Up/DownN/Baseline all call this before
// touching the database further, so drift is caught before anything else
// runs.
func verify(migrations []Migration, applied []Record) error {
	byVersion := make(map[string]Migration, len(migrations))
	for _, m := range migrations {
		byVersion[m.Version] = m
	}

	var errs []error
	for _, r := range applied {
		m, ok := byVersion[r.Version]
		if !ok {
			errs = append(errs, fmt.Errorf("%w: version %s", ErrMissingFile, r.Version))
			continue
		}
		if m.Checksum != r.Checksum {
			errs = append(errs, fmt.Errorf("%w: version %s (%s)", ErrChecksumMismatch, m.Version, m.Name))
		}
	}
	return errors.Join(errs...)
}
