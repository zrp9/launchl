package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/uptrace/bun/migrate"
	"github.com/urfave/cli/v2"
	"github.com/zrp9/launchl/cmd/migrator/migrations"
	"github.com/zrp9/launchl/internal/config"
	"github.com/zrp9/launchl/internal/database/store"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("no .env file found")
	}
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config %v", err)
	}

	dbcon, err := store.DBCon(cfg.Database)
	if err != nil {
		log.Fatalf("could not connect to database")
		return
	}

	dbStore := store.NewBuilder().SetDB(dbcon).SetBunDB().Build()
	dbStore.BnDB().AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithEnabled(false),
		bundebug.FromEnv(),
	))

	app := &cli.App{
		Name: "migrate",
		Commands: []*cli.Command{
			newMigrationCmd(migrate.NewMigrator(dbStore.BnDB(), migrations.New(), migrate.WithMarkAppliedOnSuccess(true))),
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatalf("error running migrate cli %v", err)
	}
}

func newMigrationCmd(migrator *migrate.Migrator) *cli.Command {
	return &cli.Command{
		Name:  "db",
		Usage: "database migrations",
		Subcommands: []*cli.Command{
			{
				Name:  "init",
				Usage: "create migtations table",
				Action: func(ctx *cli.Context) error {
					return migrator.Init(ctx.Context)
				},
			},
			{
				Name:  "up",
				Usage: "run up migration",
				Action: func(ctx *cli.Context) error {
					if err := migrator.Lock(ctx.Context); err != nil {
						return err
					}

					defer migrator.Unlock(ctx.Context) //nolint:errcheck
					group, err := migrator.Migrate(ctx.Context)
					if err != nil {
						return err
					}

					if group.IsZero() {
						fmt.Printf("database is up to date, no new migrations")
						return nil
					}
					fmt.Printf("successfully migrated to %s\n", group)
					return nil
				},
			},
			{
				Name:  "down",
				Usage: "run down migration",
				Action: func(ctx *cli.Context) error {
					if err := migrator.Lock(ctx.Context); err != nil {
						return err
					}

					defer migrator.Unlock(ctx.Context) //nolint:errcheck
					group, err := migrator.Rollback(ctx.Context)
					if err != nil {
						return err
					}

					if group.IsZero() {
						fmt.Printf("there are no groups to rollback")
						return nil
					}

					fmt.Printf("successfully rolled back to %s\n", group)
					return nil
				},
			},
			{
				Name:  "lock",
				Usage: "lock migrations",
				Action: func(ctx *cli.Context) error {
					return migrator.Lock(ctx.Context)
				},
			},
			{
				Name:  "unlock",
				Usage: "unlock migrations",
				Action: func(ctx *cli.Context) error {
					return migrator.Unlock(ctx.Context)
				},
			},

			{
				Name:  "make_sql",
				Usage: "create up and down sql migrations",
				Action: func(ctx *cli.Context) error {
					name := strings.Join(ctx.Args().Slice(), "_")
					files, err := migrator.CreateSQLMigrations(ctx.Context, name)
					if err != nil {
						return err
					}
					for _, f := range files {
						fmt.Printf("created migration %s (%s)\n", f.Name, f.Path)
					}
					return nil
				},
			},
			{
				Name:  "make_xsql",
				Usage: "create up and down transactional sql migrations",
				Action: func(ctx *cli.Context) error {
					name := strings.Join(ctx.Args().Slice(), "_")
					files, err := migrator.CreateTxSQLMigrations(ctx.Context, name)
					if err != nil {
						return err
					}

					for _, mf := range files {
						fmt.Printf("created transaction migrations %s (%s)\n", mf.Name, mf.Path)
					}
					return nil
				},
			},
			{
				Name:  "make_go",
				Usage: "creates Go migration",
				Action: func(ctx *cli.Context) error {
					name := strings.Join(ctx.Args().Slice(), "_")
					mf, err := migrator.CreateGoMigration(ctx.Context, name)
					if err != nil {
						return err
					}

					fmt.Printf("created migration %s (%s)\n", mf.Name, mf.Path)
					return nil
				},
			},
			{
				Name:  "status",
				Usage: "print migration status",
				Action: func(ctx *cli.Context) error {
					ms, err := migrator.MigrationsWithStatus(ctx.Context)
					if err != nil {
						return err
					}

					fmt.Printf("migrations: %s\n", ms)
					fmt.Printf("unapplied migrations: %s\n", ms.Unapplied())
					fmt.Printf("last migration group applied: %s\n", ms.LastGroup())
					return nil
				},
			},
			{
				Name:  "fake_mark",
				Usage: "mark migration as applied without running it",
				Action: func(ctx *cli.Context) error {
					group, err := migrator.Migrate(ctx.Context, migrate.WithNopMigration())
					if err != nil {
						return err
					}

					if group.IsZero() {
						fmt.Printf("there are no new migrations to mark as applied\n")
						return nil
					}
					fmt.Printf("marked %s as applied", group)
					return nil
				},
			},
		},
	}
}
