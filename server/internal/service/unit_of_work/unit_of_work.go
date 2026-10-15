// Package unit_of_work defines the agnostic Unit of Work ports services use
// to make multi-repository operations atomic, with zero knowledge of what
// actually backs it (e.g. a Postgres transaction — see
// internal/infrastructure/database for that).
package unit_of_work

import "jww/internal/domain"

// Manager creates active units of work. Services depend on the manager when
// they need an atomic operation spanning more than one repository call.
type Manager interface {
	Create(ctx domain.AppContext) (UnitOfWork, error)
}

// UnitOfWork represents one active unit of atomic work. Repositories accept it
// optionally: if present, the concrete gateway uses the underlying transaction;
// otherwise it uses its normal connection handle.
type UnitOfWork interface {
	Finish(ctx domain.AppContext) error
	Restore(ctx domain.AppContext) error
	GetValue() interface{}
}
