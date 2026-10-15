// Package bootstrap seeds the fixed set of local development test accounts
// and currencies. It builds on scripts/signup's and scripts/addcurrency's
// Clients, so each row is inserted with the same gateways and services the
// server uses.
//
// The command that calls it lives in cmd/.
package bootstrap

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"jww/internal/domain"
	"jww/scripts/addcurrency"
	"jww/scripts/signup"
	"jww/utils/logger"
)

var log = logger.New().WithComponent("scripts.bootstrap")

// testAccount is one local development login to seed. Its password is the
// username itself (see password) — these are throwaway accounts, never real
// credentials.
type testAccount struct {
	Username    string
	DisplayName string
}

// password returns the account's password: deliberately identical to its
// username, so the dev logins are trivial to remember.
func (a testAccount) password() string { return a.Username }

// testAccounts is the fixed set of accounts seeded into the local database.
var testAccounts = []testAccount{
	{Username: "julcia", DisplayName: "Julcia"},
	{Username: "wozi", DisplayName: "Wozi"},
	{Username: "baby", DisplayName: "Baby"},
}

// currencySeed is one currency to seed into the local database.
type currencySeed struct {
	Code   string
	Name   string
	IsBase bool
}

// currencies is the fixed set of currencies seeded into the local database.
// PLN is the base currency.
var currencies = []currencySeed{
	{Code: "PLN", Name: "Polish Złoty", IsBase: true},
	{Code: "USD", Name: "US Dollar"},
	{Code: "EUR", Name: "Euro"},
}

// Seed creates every test account and currency that does not exist yet,
// leaving existing ones untouched so re-running is safe.
func Seed(ctx context.Context, pool *pgxpool.Pool) error {
	userClient := signup.NewClient(pool)
	currencyClient := addcurrency.NewClient(pool)

	var failures int
	for _, account := range testAccounts {
		created, skipped, err := ensureAccount(ctx, userClient, account)
		if err != nil {
			failures++
			log.Slog().Error("account failed",
				slog.String("username", account.Username),
				slog.String("error", err.Error()),
			)
			continue
		}
		if skipped {
			log.Slog().Info("account already exists, skipping",
				slog.String("user_id", created.ID.String()),
				slog.String("username", created.Username),
			)
			continue
		}
		log.Slog().Info("account created",
			slog.String("user_id", created.ID.String()),
			slog.String("username", created.Username),
		)
	}

	for _, c := range currencies {
		created, skipped, err := ensureCurrency(ctx, currencyClient, c)
		if err != nil {
			failures++
			log.Slog().Error("currency failed",
				slog.String("code", c.Code),
				slog.String("error", err.Error()),
			)
			continue
		}
		if skipped {
			log.Slog().Info("currency already exists, skipping",
				slog.String("code", created.Code),
			)
			continue
		}
		log.Slog().Info("currency created",
			slog.String("code", created.Code),
			slog.String("name", created.Name),
			slog.Bool("is_base", created.IsBase),
		)
	}

	if failures > 0 {
		return fmt.Errorf("%d of %d seed items failed", failures, len(testAccounts)+len(currencies))
	}
	return nil
}

// ensureAccount creates account if it does not exist yet, returning the stored
// user and whether it was skipped as a pre-existing row.
func ensureAccount(ctx context.Context, client *signup.Client, account testAccount) (*domain.User, bool, error) {
	appCtx := newAppContext(ctx)

	existing, err := client.GetByUsername(appCtx, account.Username)
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		return existing, true, nil
	}

	created, err := client.Create(appCtx, account.Username, account.password(), account.DisplayName)
	if err != nil {
		return nil, false, err
	}
	return created, false, nil
}

// ensureCurrency creates c if it does not exist yet, returning the stored
// currency and whether it was skipped as a pre-existing row.
func ensureCurrency(ctx context.Context, client *addcurrency.Client, c currencySeed) (*domain.Currency, bool, error) {
	appCtx := newAppContext(ctx)

	existing, err := client.GetByCode(appCtx, c.Code)
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		return existing, true, nil
	}

	created, err := client.Create(appCtx, c.Code, c.Name, c.IsBase)
	if err != nil {
		return nil, false, err
	}
	return created, false, nil
}

func newAppContext(ctx context.Context) domain.AppContext {
	var appCtx domain.AppContext
	appCtx.SetContext(ctx)
	appCtx.SetFlowID(uuid.NewString())
	return appCtx
}
