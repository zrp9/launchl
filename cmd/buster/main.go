package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zrp9/launchl/internal/adapter/cache/valkaree"
	"github.com/zrp9/launchl/internal/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Printf("failed to load config %v exiting...", err)
		return
	}

	log.Printf("args %v", os.Args)
	if err := run(ctx, *cfg, os.Args[1:]); err != nil {
		log.Printf("execution error %v", err)
		return
	}
}

func run(ctx context.Context, cfg config.Config, args []string) error {
	ctx, cancel := context.WithTimeout(ctx, 5.*time.Minute)
	defer cancel()

	vclient, err := valkaree.NewClient(ctx, cfg.Valkey)
	if err != nil {
		log.Printf("could not create valkey client %v", err)
		return err
	}

	cache := valkaree.NewCache(vclient)
	if err := cache.DeleteMulti(ctx, args); err != nil {
		log.Printf("failed to bust cache %v", err)
	}
	// for _, arg := range args {
	// 	log.Printf("deleting key %v", arg)
	// 	if err := cache.Delete(ctx, arg); err != nil {
	// 		log.Printf("failed to delete key %v - %v", arg, err)
	// 		continue
	// 	}
	// }

	log.Println("cache busted...")
	return nil
}
