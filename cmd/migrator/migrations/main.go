// Package migrations embeds migrations Filesystem and provides way to run migrations from migrator
package migrations

import (
	"embed"

	"github.com/uptrace/bun/migrate"
)

var migrations = migrate.NewMigrations()

func New() *migrate.Migrations {
	return migrations
}

func init() {
	if err := migrations.DiscoverCaller(); err != nil {
		panic(err)
	}
}

//go:embed *.sql
var sqlMigrations embed.FS

func init() {
	if err := migrations.Discover(sqlMigrations); err != nil {
		panic(err)
	}
}
