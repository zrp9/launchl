// Package seeder can be used to seed db with initial data
package seeder

import (
	"database/sql"
	"log"

	"github.com/zrp9/launchl/internal/config"
	"github.com/zrp9/launchl/internal/database/store"
	"github.com/zrp9/launchl/internal/seeder"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Println("failed to load database config exiting...")
		return
	}

	conn, err := store.DBCon(cfg.Database)
	if err != nil {
		log.Printf("an erro occurred while connecting to db %v", err)
	}
	if err := run(cfg.Server, conn); err != nil {
		log.Printf("an error occurred while running server %v", err)
	}
}

func run(serverCfg config.ServerCfg, con *sql.DB) error {
	dbStore := store.NewBuilder().SetDB(con).SetBunDB().RegisterModels().Build()
	adapter := seeder.SeederFactory(dbStore)
	if err := adapter.LoadDB(); err != nil {
		return err
	}

	return nil
}
