// Command migrate is the manual CLI for the schema migrations embedded in
// jww/migrations — the same files internal/infrastructure/database.Migrate
// applies automatically at server startup. See migrations/README.md for the
// file format and migrations/README.md#cutting-over-an-existing-database for
// the one-time `baseline` step.
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"jww/internal/infrastructure/database"
	"jww/internal/infrastructure/migrator"
	"jww/migrations"
	"jww/utils/logger"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: migrate <up|down [n]|status|baseline>")
	}

	if err := logger.Configure(logger.Config{Level: "info", Pretty: true}); err != nil {
		return err
	}
	log := logger.New().WithComponent("cmd.migrate").Slog()

	cfg, err := database.LoadConfig()
	if err != nil {
		return err
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()

	switch args[0] {
	case "up":
		applied, err := migrator.Up(ctx, pool, migrations.FS, log)
		if err != nil {
			return err
		}
		if len(applied) == 0 {
			fmt.Println("database schema up to date")
		}
		return nil

	case "down":
		n := 1
		if len(args) > 1 {
			n, err = strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid step count %q: %w", args[1], err)
			}
		}
		reversed, err := migrator.DownN(ctx, pool, migrations.FS, n, log)
		if err != nil {
			return err
		}
		if len(reversed) == 0 {
			fmt.Println("nothing to reverse")
		}
		return nil

	case "status":
		entries, err := migrator.Status(ctx, pool, migrations.FS)
		if err != nil {
			return err
		}
		for _, e := range entries {
			state := "pending"
			switch {
			case e.Missing:
				state = "APPLIED (file missing!)"
			case e.Applied && e.Drifted:
				state = "APPLIED (checksum mismatch!)"
			case e.Applied:
				state = "applied"
			}
			fmt.Printf("%s_%s: %s\n", e.Migration.Version, e.Migration.Name, state)
		}
		return nil

	case "baseline":
		baselined, err := migrator.Baseline(ctx, pool, migrations.FS, log)
		if err != nil {
			return err
		}
		fmt.Printf("baselined %d migration(s)\n", len(baselined))
		return nil

	default:
		return fmt.Errorf("unknown command %q (want up|down [n]|status|baseline)", args[0])
	}
}
