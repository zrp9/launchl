// Package seeder can be used to seed db with initial data
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zrp9/launchl/internal/adapter/cache/valkaree"
	"github.com/zrp9/launchl/internal/adapter/repo/pgsql"
	"github.com/zrp9/launchl/internal/config"
	"github.com/zrp9/launchl/internal/domain/service"
	"github.com/zrp9/launchl/internal/seeder"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Println("failed to load database config exiting...")
		return
	}

	fmt.Println("Args:", os.Args)
	if len(os.Args) > 1 {
		fmt.Println("First argument:", os.Args[1])
	}

	// only pass in args after the first index
	if err := run(ctx, *cfg, os.Args[1:]); err != nil {
		log.Printf("an error occurred while running server %v", err)
	}
}

func run(ctx context.Context, cfg config.Config, args []string) error {
	ctx, cancel := context.WithTimeout(ctx, 5.*time.Minute)
	defer cancel()

	// TODO: add connectRetry and close to db conn
	conn, err := pgsql.DBConWithRetry(cfg.Database)
	if err != nil {
		log.Printf("an erro occurred while connecting to db %v", err)
		return err
	}

	vclient, err := valkaree.NewClient(ctx, cfg.Valkey)
	if err != nil {
		log.Printf("an error occurred creating valkey client %v", err)
		return err
	}

	dbStore := pgsql.NewBuilder().SetDB(conn).SetBunDB().RegisterModels().Build()
	cache := valkaree.NewCache(vclient)

	appRepo := pgsql.NewAppRepo(dbStore)
	appService := service.NewAppService(appRepo, cache)
	log.Printf("seeder args %v", args)
	adapter := seeder.SeederFactory(appService, args)
	if err := adapter.LoadDBData(); err != nil {
		return err
	}

	return nil
}
