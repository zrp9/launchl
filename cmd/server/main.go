package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/zrp9/launchl/internal/api"
	"github.com/zrp9/launchl/internal/app"
	"github.com/zrp9/launchl/internal/config"
	"github.com/zrp9/launchl/internal/crane"
	"github.com/zrp9/launchl/internal/database/store"
)

func main() {
	fmt.Println("running on 8090")
	services := []string{"user"}
	cfg, err := config.Load()
	if err != nil {
		log.Println("failed to load database config exiting...")
		return
	}

	conn, err := store.DBCon(cfg.Database)
	if err != nil {
		log.Printf("an erro occurred while connecting to db %v", err)
	}
	if err := run(cfg.Server, conn, services); err != nil {
		log.Printf("an error occurred while running server %v", err)
	}
}

func run(serverCfg config.ServerCfg, con *sql.DB, services []string) error {
	logger := crane.DefaultLogger
	dbStore := store.NewBuilder().SetDB(con).SetBunDB().RegisterModels().Build()
	// userRepo := urepo.New(dbStore)
	// usrService := usr.New(userRepo)
	// userApi := usr.Initialize(usrService, logger)

	container := app.New(dbStore, logger)
	if err := container.RegisterServices(services); err != nil {
		logger.MustDebugErr(err)
		return err
	}
	server := api.NewServer(serverCfg, container.Endpoints())

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.MustDebugErr(err)
		return err
	}
	return nil
}
