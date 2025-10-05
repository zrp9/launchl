// Package seeder can be used to seed db with initial data
package main

import (
	"context"
	"database/sql"
	"log"

	"github.com/zrp9/launchl/internal/adapter/cache/valkaree"
	"github.com/zrp9/launchl/internal/adapter/repo/pgsql"
	"github.com/zrp9/launchl/internal/config"
	"github.com/zrp9/launchl/internal/database/store"
	"github.com/zrp9/launchl/internal/domain/service"
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
	if err := run(conn); err != nil {
		log.Printf("an error occurred while running server %v", err)
	}
}

func run(con *sql.DB) error {
	dbStore := store.NewBuilder().SetDB(con).SetBunDB().RegisterModels().Build()
	cache, err := valkaree.NewCache(context.TODO())
	if err != nil {
		return err
	}
	appRepo := pgsql.NewAppRepo(dbStore)
	appService := service.NewAppService(appRepo, cache)
	adapter := seeder.SeederFactory(appService)
	if err := adapter.LoadDBData(); err != nil {
		return err
	}

	return nil
}
