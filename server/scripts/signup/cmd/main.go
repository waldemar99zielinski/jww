// Command signup inserts a user straight into the database, bypassing the
// HTTP API. It is a maintenance script rather than part of cmd/server: it is
// built and run on demand, e.g.
//
//	go run ./scripts/signup/cmd -username alice -password s3cret -database-url postgres://...
//
// With -pgen it generates a secure password instead of taking -password, and
// logs the generated value after the account is created:
//
//	go run ./scripts/signup/cmd -username alice -pgen -database-url postgres://...
//
// Each value may instead come from an environment variable (see parseParams);
// the flag wins when both are set, and a value that neither supplies is an
// error. The insert goes through the signup package, which builds the same
// gateway and Argon2id hasher the server uses, so the resulting row logs in
// through /v1/auth/login like any other signup.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"jww/internal/domain"
	"jww/internal/infrastructure/database"
	"jww/scripts/signup"
	"jww/utils/logger"
)

var log = logger.New().WithComponent("scripts.signup")

func main() {
	params, err := parseParams(os.Args[1:], os.Getenv)
	if err != nil {
		log.Slog().Error("signup failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := run(params); err != nil {
		log.Slog().Error("signup failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

// params is everything run needs: the database to talk to and the account to
// create.
type params struct {
	databaseURL string
	username    string
	password    string
	displayName string
	pgen        bool
}

// parseParams resolves every value from a command-line flag first, then the
// matching environment variable, and fails when neither supplies a required
// one.
func parseParams(args []string, getenv func(string) string) (params, error) {
	fs := flag.NewFlagSet("signup", flag.ContinueOnError)
	var p params

	fs.StringVar(&p.databaseURL, "database-url", getenv("DATABASE_URL"), "Postgres DSN (or DATABASE_URL)")
	fs.StringVar(&p.username, "username", getenv("SIGNUP_USERNAME"), "account username (or SIGNUP_USERNAME)")
	fs.StringVar(&p.password, "password", getenv("SIGNUP_PASSWORD"), "account password (or SIGNUP_PASSWORD)")
	fs.StringVar(&p.displayName, "display-name", getenv("SIGNUP_DISPLAY_NAME"), "display name (or SIGNUP_DISPLAY_NAME; defaults to username)")
	fs.BoolVar(&p.pgen, "pgen", false, "generate a secure password instead of taking -password (flag only, off by default)")
	if err := fs.Parse(args); err != nil {
		return p, err
	}

	if p.pgen && p.password != "" {
		return p, fmt.Errorf("-pgen cannot be combined with -password or SIGNUP_PASSWORD")
	}

	var missing []string
	if p.databaseURL == "" {
		missing = append(missing, "-database-url or DATABASE_URL")
	}
	if p.username == "" {
		missing = append(missing, "-username or SIGNUP_USERNAME")
	}
	if !p.pgen && p.password == "" {
		missing = append(missing, "-password or SIGNUP_PASSWORD (or use -pgen)")
	}
	if len(missing) > 0 {
		return p, fmt.Errorf("missing required value(s): %s", strings.Join(missing, ", "))
	}

	if p.displayName == "" {
		p.displayName = p.username
	}
	return p, nil
}

func run(p params) error {
	password := p.password
	if p.pgen {
		generated, err := signup.GeneratePassword()
		if err != nil {
			return err
		}
		password = generated
	}

	loggerConfig, err := logger.LoadConfig()
	if err != nil {
		return err
	}
	if err := logger.Configure(loggerConfig); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, database.Config{URL: p.databaseURL})
	if err != nil {
		return err
	}
	defer pool.Close()

	var appCtx domain.AppContext
	appCtx.SetContext(ctx)
	appCtx.SetFlowID(uuid.NewString())

	created, err := signup.NewClient(pool).Create(appCtx, p.username, password, p.displayName)
	if err != nil {
		return err
	}

	log.Slog().Info("user created",
		slog.String("user_id", created.ID.String()),
		slog.String("username", created.Username),
	)
	if p.pgen {
		log.Slog().Info("generated password",
			slog.String("username", created.Username),
			slog.String("password", password),
		)
	}
	return nil
}
