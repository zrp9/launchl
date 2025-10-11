package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/zrp9/launchl/internal/adapter/log/crane"
	"github.com/zrp9/launchl/internal/adapter/repo/pgsql"
	"github.com/zrp9/launchl/internal/app"
	"github.com/zrp9/launchl/internal/config"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic: %v\n%s", r, debug.Stack())
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	// handle termination singals
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	services := []string{"user", "survey", "app"}
	cfg, err := config.Load()
	fmt.Printf("running on %v\n", net.JoinHostPort("", cfg.Server.Port))
	if err != nil {
		log.Println("failed to load database config exiting...")
		return
	}

	conn, err := pgsql.DBConWithRetry(cfg.Database)
	if err != nil {
		log.Printf("an erro occurred while connecting to db %v", err)
		return
	}

	logger := crane.DefaultLogger
	dbStore := pgsql.NewBuilder().SetDB(conn).SetBunDB().RegisterModels().Build()
	container := app.New(dbStore, logger)
	// TODO this might have to be updated
	ctx := context.TODO()
	if err := container.RegisterServices(ctx, services); err != nil {
		log.Fatalf("register services: %v", err) // non-zero exit, you'll see the reason
		return
	}

	mux := http.NewServeMux()
	log.Printf("MUX in main create:  %p", mux)
	server := app.NewServer(cfg.Server, mux, container.Handlers())
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
		return
	}

	log.Print("server stopped gracefully")
}
