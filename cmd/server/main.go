package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zrp9/launchl/internal/adapter/log/crane"
	"github.com/zrp9/launchl/internal/app"
	"github.com/zrp9/launchl/internal/config"
	"github.com/zrp9/launchl/internal/database/store"
)

func main() {
	fmt.Println("running on 8090")
	stop := make(chan os.Signal, 1)
	// handle termination singals
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	services := []string{"user", "survey", "app"}
	cfg, err := config.Load()
	if err != nil {
		log.Println("failed to load database config exiting...")
		return
	}

	conn, err := store.DBCon(cfg.Database)
	if err != nil {
		log.Printf("an erro occurred while connecting to db %v", err)
	}

	logger := crane.DefaultLogger
	dbStore := store.NewBuilder().SetDB(conn).SetBunDB().RegisterModels().Build()
	container := app.New(dbStore, logger)
	// TODO this might have to be updated
	ctx := context.TODO()
	if err := container.RegisterServices(ctx, services); err != nil {
		logger.MustDebugErr(err)
	}
	server := app.NewServer(cfg.Server, container.Handlers())

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.MustFatal(err.Error())
		}
	}()

	// wait for go routine running server to stop
	<-stop
	log.Println("shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("force shutdown %v", err)
		logger.MustFatal(err.Error())
	}

	log.Print("server stopped gracefully")
}
